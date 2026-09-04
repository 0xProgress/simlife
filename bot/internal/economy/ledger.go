package economy

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

var (
	ErrInvalidAmount       = errors.New("transaction amount must be greater than zero")
	ErrInsufficientFunds   = errors.New("source account has insufficient funds")
	ErrAccountNotFound     = errors.New("one or both accounts do not exist")
	ErrSameAccount         = errors.New("source and destination accounts cannot be the same")
	ErrSerializationFailed = errors.New("transaction failed due to concurrent modification, please retry")
)

// Ledger is the authoritative double-entry bookkeeping engine for Simlife.
type Ledger struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
	log     zerolog.Logger
}

// NewLedger initializes the financial ledger.
func NewLedger(pool *pgxpool.Pool, queries *sqlc.Queries, log zerolog.Logger) *Ledger {
	return &Ledger{
		pool:    pool,
		queries: queries,
		log:     log.With().Str("component", "ledger").Logger(),
	}
}

// Transfer is the primary public method for moving funds. It wraps PostTransaction
// with exponential backoff retry logic for serialization failures.
func (l *Ledger) Transfer(ctx context.Context, sourceID, destID string, amount decimal.Decimal, txType, playerID, description string) error {
	maxRetries := 5
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := l.postTransaction(ctx, sourceID, destID, amount, txType, playerID, description)
		if err == nil {
			return nil
		}
		if errors.Is(err, ErrSerializationFailed) && attempt < maxRetries {
			// Exponential backoff: 10ms, 40ms, 90ms, 160ms
			backoff := time.Duration(attempt*attempt) * 10 * time.Millisecond
			l.log.Warn().Err(err).Int("attempt", attempt).Dur("backoff", backoff).Msg("serialization failure, retrying")
			time.Sleep(backoff)
			continue
		}
		return fmt.Errorf("ledger transfer failed: %w", err)
	}
	return ErrSerializationFailed
}

// postTransaction atomically posts a double-entry transaction.
func (l *Ledger) postTransaction(ctx context.Context, sourceID, destID string, amount decimal.Decimal, txType, playerID, description string) error {
	start := time.Now()

	if amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}
	if sourceID == destID {
		return ErrSameAccount
	}

	// Begin SERIALIZABLE transaction
	tx, err := l.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("failed to begin serializable transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := l.queries.WithTx(tx)

	// Lock accounts in consistent order to prevent deadlocks
	ids := []string{sourceID, destID}
	if bytes.Compare([]byte(ids[0]), []byte(ids[1])) > 0 {
		ids[0], ids[1] = ids[1], ids[0]
	}

	accounts, err := qtx.LockAccountsForUpdate(ctx, ids)
	if err != nil {
		return fmt.Errorf("failed to lock accounts: %w", err)
	}
	if len(accounts) != 2 {
		return ErrAccountNotFound
	}

	// Validate sufficient funds (skip for Treasury minting if source is Treasury)
	if sourceID != "TREASURY" { // Assuming Treasury has a known ID or we check account type
		balance, err := qtx.GetAccountBalance(ctx, sourceID)
		if err != nil {
			return fmt.Errorf("failed to compute source balance: %w", err)
		}
		currentBal := numericToDecimal(balance)
		if currentBal.LessThan(amount) {
			return ErrInsufficientFunds
		}
	}

	// Post DEBIT
	_, err = qtx.PostTransaction(ctx, sqlc.PostTransactionParams{
		AccountID:       sourceID,
		Amount:          decimalToNumeric(amount),
		EntryType:       sqlc.EntryTypeDEBIT,
		TransactionType: txType,
		ReferenceID:     pgtype.Text{String: playerID, Valid: playerID != ""},
		Description:     pgtype.Text{String: description, Valid: description != ""},
	})
	if err != nil {
		return fmt.Errorf("failed to post debit: %w", err)
	}

	// Post CREDIT
	_, err = qtx.PostTransaction(ctx, sqlc.PostTransactionParams{
		AccountID:       destID,
		Amount:          decimalToNumeric(amount),
		EntryType:       sqlc.EntryTypeCREDIT,
		TransactionType: txType,
		ReferenceID:     pgtype.Text{String: playerID, Valid: playerID != ""},
		Description:     pgtype.Text{String: description, Valid: description != ""},
	})
	if err != nil {
		return fmt.Errorf("failed to post credit: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		if isSerializationFailure(err) {
			return ErrSerializationFailed
		}
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	l.log.Info().
		Str("type", txType).
		Str("source_account", sourceID).
		Str("dest_account", destID).
		Str("amount", amount.String()).
		Dur("duration_ms", time.Since(start)).
		Msg("transaction_posted_successfully")

	return nil
}

func isSerializationFailure(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "40001"
	}
	return false
}

func decimalToNumeric(d decimal.Decimal) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Scan(d.String())
	return n
}

func numericToDecimal(n pgtype.Numeric) decimal.Decimal {
	if !n.Valid {
		return decimal.Zero
	}
	d, err := decimal.NewFromString(n.String())
	if err != nil {
		return decimal.Zero
	}
	return d
}