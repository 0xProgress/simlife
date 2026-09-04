package world

import (
	"context"
	"errors"
	"fmt"

	sqlc "github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

var (
	ErrPlotNotFound      = errors.New("plot not found")
	ErrPlotAlreadyOwned  = errors.New("plot is already owned")
	ErrPlotNotOwned      = errors.New("plot is not owned")
	ErrUnauthorized      = errors.New("you do not own this plot")
	ErrMaxDevLevel       = errors.New("plot is at maximum development level")
	ErrGovZoneRestricted = errors.New("government zone plots cannot be privately owned")
)

// PlotManager handles all operations on individual city plots.
type PlotManager struct {
	queries *sqlc.Queries
	log     zerolog.Logger
}

// NewPlotManager initializes the plot manager.
func NewPlotManager(q *sqlc.Queries, log zerolog.Logger) *PlotManager {
	return &PlotManager{
		queries: q,
		log:     log.With().Str("component", "world.plot").Logger(),
	}
}

// Plot represents a single city plot with all its state.
type Plot struct {
	ID               string
	PlotX            int32
	PlotY            int32
	ZoneType         ZoneType
	OwnerPlayerID    string // Empty string if unowned
	DevelopmentLevel int
	AssessedValue    decimal.Decimal
	LastTaxPayment   pgtype.Timestamptz
}

// GetPlot fetches a single plot by ID.
func (p *PlotManager) GetPlot(ctx context.Context, plotID string) (*Plot, error) {
	row, err := p.queries.GetPlotByID(ctx, plotID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlotNotFound
		}
		return nil, fmt.Errorf("failed to fetch plot: %w", err)
	}

	return plotFromRow(row), nil
}

// GetAvailablePlots fetches all unowned plots, optionally filtered by zone type.
func (p *PlotManager) GetAvailablePlots(ctx context.Context, zoneType ZoneType, limit int32) ([]*Plot, error) {
	var rows []sqlc.Plot
	var err error

	if zoneType == "" {
		rows, err = p.queries.GetAvailablePlots(ctx, limit)
	} else {
		rows, err = p.queries.GetAvailablePlotsByZone(ctx, sqlc.GetAvailablePlotsByZoneParams{
			ZoneType: string(zoneType),
			Limit:    limit,
		})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch available plots: %w", err)
	}

	plots := make([]*Plot, 0, len(rows))
	for _, row := range rows {
		plots = append(plots, plotFromRow(row))
	}
	return plots, nil
}

// GetPlotsByOwner fetches all plots owned by a specific player.
func (p *PlotManager) GetPlotsByOwner(ctx context.Context, playerID string) ([]*Plot, error) {
	rows, err := p.queries.GetPlotsByOwner(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch player plots: %w", err)
	}

	plots := make([]*Plot, 0, len(rows))
	for _, row := range rows {
		plots = append(plots, plotFromRow(row))
	}
	return plots, nil
}

// PurchasePlot transfers ownership of an unowned plot to a player.
// This does NOT handle payment — the caller must execute the ledger transfer first.
func (p *PlotManager) PurchasePlot(ctx context.Context, plotID, playerID string) error {
	log := logger.FromContext(ctx, "world.plot")

	plot, err := p.GetPlot(ctx, plotID)
	if err != nil {
		return err
	}

	if plot.ZoneType == ZoneGovernment {
		return ErrGovZoneRestricted
	}

	if plot.OwnerPlayerID != "" {
		return ErrPlotAlreadyOwned
	}

	err = p.queries.TransferPlotOwnership(ctx, sqlc.TransferPlotOwnershipParams{
		ID:            plotID,
		OwnerPlayerID: playerID,
	})
	if err != nil {
		return fmt.Errorf("failed to transfer plot ownership: %w", err)
	}

	log.Info().
		Str("plot_id", plotID).
		Str("player_id", playerID).
		Str("zone", string(plot.ZoneType)).
		Msg("plot purchased successfully")

	return nil
}

// SellPlot releases ownership of a player-owned plot back to the city.
// This does NOT handle payment — the caller must execute the ledger transfer first.
func (p *PlotManager) SellPlot(ctx context.Context, plotID, playerID string) error {
	log := logger.FromContext(ctx, "world.plot")

	plot, err := p.GetPlot(ctx, plotID)
	if err != nil {
		return err
	}

	if plot.OwnerPlayerID != playerID {
		return ErrUnauthorized
	}

	err = p.queries.ReleasePlotOwnership(ctx, plotID)
	if err != nil {
		return fmt.Errorf("failed to release plot ownership: %w", err)
	}

	log.Info().
		Str("plot_id", plotID).
		Str("player_id", playerID).
		Msg("plot sold successfully")

	return nil
}

// DevelopPlot increases the development level of a player-owned plot by 1.
// Returns the new assessed value after development.
func (p *PlotManager) DevelopPlot(ctx context.Context, plotID, playerID string) (decimal.Decimal, error) {
	log := logger.FromContext(ctx, "world.plot")

	plot, err := p.GetPlot(ctx, plotID)
	if err != nil {
		return decimal.Zero, err
	}

	if plot.OwnerPlayerID != playerID {
		return decimal.Zero, ErrUnauthorized
	}

	zoneCfg, err := GetZoneConfig(plot.ZoneType)
	if err != nil {
		return decimal.Zero, err
	}

	if plot.DevelopmentLevel >= zoneCfg.MaxDevelopmentLevel {
		return decimal.Zero, ErrMaxDevLevel
	}

	err = p.queries.UpgradePlotDevelopment(ctx, plotID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to upgrade plot development: %w", err)
	}

	newValue := CalculateBasePropertyValue(plot.ZoneType, plot.DevelopmentLevel+1)

	log.Info().
		Str("plot_id", plotID).
		Int("new_level", plot.DevelopmentLevel+1).
		Str("new_value", newValue.String()).
		Msg("plot developed successfully")

	return newValue, nil
}

// GetPlotSalePrice calculates the current market price for a plot.
// This factors in zone, development level, and a 10% depreciation on resale.
func (p *PlotManager) GetPlotSalePrice(ctx context.Context, plotID string) (decimal.Decimal, error) {
	plot, err := p.GetPlot(ctx, plotID)
	if err != nil {
		return decimal.Zero, err
	}

	assessedValue := plot.AssessedValue
	if assessedValue.IsZero() {
		assessedValue = CalculateBasePropertyValue(plot.ZoneType, plot.DevelopmentLevel)
	}

	// 10% depreciation on resale to discourage flipping
	salePrice := assessedValue.Mul(decimal.NewFromFloat(0.90)).Truncate(0)
	return salePrice, nil
}

// GetPlotPurchasePrice calculates the purchase price for an unowned plot.
func (p *PlotManager) GetPlotPurchasePrice(ctx context.Context, plotID string) (decimal.Decimal, error) {
	plot, err := p.GetPlot(ctx, plotID)
	if err != nil {
		return decimal.Zero, err
	}

	if plot.OwnerPlayerID != "" {
		return decimal.Zero, ErrPlotAlreadyOwned
	}

	basePrice := CalculateBasePropertyValue(plot.ZoneType, plot.DevelopmentLevel)
	return basePrice, nil
}

// plotFromRow converts a sqlc Plot row to our domain Plot struct.
func plotFromRow(row sqlc.Plot) *Plot {
	return &Plot{
		ID:               row.ID,
		PlotX:            row.PlotX,
		PlotY:            row.PlotY,
		ZoneType:         ZoneType(row.ZoneType),
		OwnerPlayerID:    row.OwnerPlayerID.String,
		DevelopmentLevel: int(row.DevelopmentLevel),
		AssessedValue:    numericToDecimal(row.AssessedValue),
		LastTaxPayment:   row.LastTaxPayment,
	}
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