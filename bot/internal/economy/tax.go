package economy

import (
	"context"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// TaxCalculator handles all tax mechanics.
// Layer 1: Transaction tax on P2P transfers.
// Layer 5: Property tax.
// Layer 7: Configurable rates (future).
type TaxCalculator struct {
	queries *sqlc.Queries
	ledger  *Ledger
	log     zerolog.Logger
}

// NewTaxCalculator initializes the tax calculator.
func NewTaxCalculator(q *sqlc.Queries, l *Ledger, log zerolog.Logger) *TaxCalculator {
	return &TaxCalculator{
		queries: q,
		ledger:  l,
		log:     log.With().Str("component", "tax").Logger(),
	}
}

// CalculateTransactionTax computes Layer 1 transaction tax (2% flat rate).
// It returns the tax amount to be routed to the city treasury as a money sink.
func (t *TaxCalculator) CalculateTransactionTax(amount decimal.Decimal) decimal.Decimal {
	taxRate := decimal.NewFromFloat(0.02)
	return amount.Mul(taxRate).Truncate(0) // Round down to whole number for simplicity
}

// CollectPropertyTaxes executes Layer 5 property tax collection during settlement.
// It iterates over all owned properties, calculates the tax based on assessed value,
// and posts a Ledger transaction from the owner's WALLET to the TREASURY.
func (t *TaxCalculator) CollectPropertyTaxes(ctx context.Context) error {
	log := logger.FromContext(ctx, "economy.tax")
	log.Info().Msg("starting property tax collection phase")

	// 1. Fetch all properties with an owner
	properties, err := t.queries.GetTaxableProperties(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch taxable properties: %w", err)
	}

	if len(properties) == 0 {
		log.Info().Msg("no taxable properties found")
		return nil
	}

	// 2. Fetch Treasury account
	treasury, err := t.queries.GetTreasuryAccount(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch treasury account: %w", err)
	}

	// 3. Process tax for each property
	successCount := 0
	failureCount := 0

	for _, prop := range properties {
		ownerID := prop.OwnerPlayerID.String
		assessedValue := numericToDecimal(prop.AssessedValue)

		// Tax rate: 1% of assessed value per economic cycle (e.g., month/day)
		taxRate := decimal.NewFromFloat(0.01)
		taxAmount := assessedValue.Mul(taxRate).Truncate(0)

		if taxAmount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		// Fetch owner's wallet
		wallet, err := t.queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
			PlayerID:    ownerID,
			AccountType: sqlc.AccountTypeWALLET,
		})
		if err != nil {
			log.Warn().Err(fmt.Errorf("failed to fetch wallet for player %s: %w", ownerID, err)).
				Str("property_id", prop.ID).
				Msg("skipping property tax collection")
			failureCount++
			continue
		}

		// Execute the tax collection via the Ledger (SERIALIZABLE, double-entry enforced)
		err = t.ledger.Transfer(ctx, wallet.ID, treasury.ID, taxAmount, "TAX_COLLECTION", ownerID, fmt.Sprintf("Property tax for plot %s", prop.ID))
		if err != nil {
			if err == ErrInsufficientFunds {
				log.Warn().Str("player_id", ownerID).Str("property_id", prop.ID).
					Msg("insufficient funds for property tax, marking as delinquent")
				// In Layer 5+, this would trigger a delinquency flag or eventual foreclosure
			} else {
				log.Error().Err(fmt.Errorf("failed to collect property tax: %w", err)).
					Str("player_id", ownerID).Str("property_id", prop.ID).
					Msg("property tax collection failed")
			}
			failureCount++
			continue
		}

		// Update last tax payment timestamp
		err = t.queries.UpdatePropertyTaxPayment(ctx, sqlc.UpdatePropertyTaxPaymentParams{
			ID: prop.ID,
		})
		if err != nil {
			log.Warn().Err(fmt.Errorf("failed to update tax payment timestamp: %w", err)).
				Str("property_id", prop.ID).Msg("tax payment recorded but timestamp update failed")
		}

		successCount++
	}

	log.Info().
		Int("total_properties", len(properties)).
		Int("successful_collections", successCount).
		Int("failed_collections", failureCount).
		Msg("property tax collection phase completed")

	return nil
}

// numericToDecimal safely converts pgtype.Numeric to decimal.Decimal.
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