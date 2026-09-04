package economy

import (
	"context"

	"github.com/rs/zerolog"

	db "github.com/0xProgress/simlife/bot/db/sqlc"
)

// TaxCalculator handles all tax mechanics.
// Layer 1: Transaction tax on P2P transfers.
// Layer 5: Property tax.
// Layer 7: Configurable rates (future).
type TaxCalculator struct {
	queries *db.Queries
	ledger  *Ledger
	log     zerolog.Logger
}

// NewTaxCalculator initializes the tax calculator.
func NewTaxCalculator(q *db.Queries, l *Ledger, log zerolog.Logger) *TaxCalculator {
	return &TaxCalculator{
		queries: q,
		ledger:  l,
		log:     log.With().Str("component", "tax").Logger(),
	}
}

// CalculateTransactionTax computes Layer 1 transaction tax.
func (t *TaxCalculator) CalculateTransactionTax(amount float64) float64 {
	// Layer 1: Flat 2% tax on P2P transfers flows to city treasury
	return amount * 0.02
}

// CollectPropertyTaxes executes Layer 5 property tax collection during settlement.
func (t *TaxCalculator) CollectPropertyTaxes(ctx context.Context) error {
	// TODO: Fetch all properties with owner_id IS NOT NULL
	// Calculate tax based on assessed_value
	// Post Ledger transaction from owner WALLET to TREASURY
	// Update last_tax_paid_at
	t.log.Info().Msg("property tax collection phase executed (Layer 5 stub)")
	return nil
}