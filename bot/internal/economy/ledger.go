package economy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	db "github.com/0xProgress/simlife/bot/db/sqlc"
)

var (
	ErrInvalidAmount       = errors.New("transaction amount must be greater than zero")
	ErrInsufficientFunds   = errors.New("source account has insufficient funds")
	ErrAccountNotFound     = errors.New("one or both accounts do not exist")
	ErrSameAccount         = errors.New("source and destination accounts cannot be the same")
	ErrSerializationFailed = errors.New("transaction failed due to concurrent modification, please retry")
)

// Ledger is the authoritative double-entry bookkeeping engine for Simlife.
// All financial state mutations must flow through this component. No other package
// is permitted to write to the financial ledger.
type Ledger struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	log     zerolog.Logger
}

// NewLedger initializes the financial ledger with database connections and a logger.
func NewLedger(pool *pgxpool.Pool, queries *db.Queries, log zerolog.Logger) *Ledger {
	return &Ledger{
		pool:    pool,
		queries: queries,
		log:     log.With().Str("component", "ledger").Logger(),
	}
}

// PostTransactionParams defines the inputs required to post a financial transaction.
type PostTransactionParams struct {
	SourceAccountID uuid.UUID
	DestAccountID   uuid.UUID
	Amount          float64 // Note: Use shopspring/decimal in production for exact precision
	Type            db.TransactionType
	Description     string
}

// PostTransaction atomically posts a double-entry transaction to the ledger.
// It enforces SERIALIZABLE isolation, validates account existence and sufficient funds,
// and guarantees that exactly one DEBIT and one CREDIT entry are inserted.
func (l *Ledger) PostTransaction(ctx context.Context, params PostTransactionParams) (uuid.UUID, error) {
	start := time.Now()

	// 1. Strict Input Validation
	if params.Amount <= 0 {
		return uuid.Nil, ErrInvalidAmount
	}
	if params.SourceAccountID == params.DestAccountID {
		return uuid.Nil, ErrSameAccount
	}

	// 2. Begin SERIALIZABLE Transaction
	// This isolation level prevents write skew and phantom reads, ensuring
	// that concurrent balance checks and inserts remain mathematically consistent.
	tx, err := l.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to begin serializable transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 3. Lock Accounts & Verify Existence
	// We lock both accounts using SELECT ... FOR UPDATE to serialize concurrent
	// modifications to the same accounts. We sort the UUIDs to enforce a consistent
	// lock ordering, which completely prevents deadlocks when two transactions
	// attempt to transfer funds between the same two accounts in opposite directions.
	ids := []uuid.UUID{params.SourceAccountID, params.DestAccountID}
	if bytes.Compare(ids[0][:], ids[1][:]) > 0 {
		ids[0], ids[1] = ids[1], ids[0]
	}

	rows, err := tx.Query(ctx, "SELECT id FROM accounts WHERE id = ANY($1) FOR UPDATE", ids)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to lock accounts: %w", err)
	}
	defer rows.Close()

	foundCount := 0
	for rows.Next() {
		foundCount++
	}
	if foundCount != 2 {
		return uuid.Nil, ErrAccountNotFound
	}

	// 4. Validate Sufficient Funds
	qtx := l.queries.WithTx(tx)
	
	balance, err := qtx.GetAccountBalance(ctx, params.SourceAccountID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to compute source account balance: %w", err)
	}

	// TODO: Exempt TREASURY accounts from balance checks to allow money supply minting
	// Cast balance to float64 in case sqlc inferred int32 from the COALESCE(..., 0) literal
	if float64(balance) < params.Amount {
		return uuid.Nil, ErrInsufficientFunds
	}

	// 5. Generate Shared Reference ID
	// This UUID correlates the DEBIT and CREDIT halves of the double-entry pair.
	refID := uuid.New()

	// Convert float64 to pgtype.Numeric for sqlc generated structs
	var pgAmount pgtype.Numeric
	if err := pgAmount.Scan(params.Amount); err != nil {
		return uuid.Nil, fmt.Errorf("failed to convert amount to pgtype.Numeric: %w", err)
	}

	// Prepare pgx specific nullable types
	pgRefID := pgtype.UUID{Bytes: refID, Valid: true}
	pgDesc := pgtype.Text{String: params.Description, Valid: params.Description != ""}

	// 6. Post DEBIT Entry (Source Account)
	_, err = qtx.PostTransaction(ctx, db.PostTransactionParams{
		AccountID:       params.SourceAccountID,
		Amount:          pgAmount,
		EntryType:       db.EntryTypeDEBIT,
		TransactionType: params.Type,
		ReferenceID:     pgRefID,
		Description:     pgDesc,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to post debit entry: %w", err)
	}

	// 7. Post CREDIT Entry (Destination Account)
	_, err = qtx.PostTransaction(ctx, db.PostTransactionParams{
		AccountID:       params.DestAccountID,
		Amount:          pgAmount,
		EntryType:       db.EntryTypeCREDIT,
		TransactionType: params.Type,
		ReferenceID:     pgRefID,
		Description:     pgDesc,
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to post credit entry: %w", err)
	}

	// 8. Commit Transaction
	if err := tx.Commit(ctx); err != nil {
		if isSerializationFailure(err) {
			return uuid.Nil, ErrSerializationFailed
		}
		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 9. Structured Logging
	l.log.Info().
		Str("transaction_id", refID.String()).
		Str("type", string(params.Type)).
		Str("source_account", params.SourceAccountID.String()).
		Str("dest_account", params.DestAccountID.String()).
		Float64("amount", params.Amount).
		Dur("duration", time.Since(start)).
		Msg("transaction_posted_successfully")

	return refID, nil
}

// isSerializationFailure checks if a PostgreSQL error is a serialization failure (SQLSTATE 40001).
// Callers should retry the transaction when this error is encountered.
func isSerializationFailure(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40001"
	}
	return false
}