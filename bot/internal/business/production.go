package business

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// ProductionManager manages the production cycle during settlement.
type ProductionManager struct {
	queries *sqlc.Queries
	log     zerolog.Logger
}

// NewProductionManager initializes the production manager.
func NewProductionManager(q *sqlc.Queries, log zerolog.Logger) *ProductionManager {
	return &ProductionManager{
		queries: q,
		log:     log.With().Str("component", "business_production").Logger(),
	}
}

// ProcessProduction runs the daily production cycle for all active businesses.
func (p *ProductionManager) ProcessProduction(ctx context.Context) error {
	log := logger.FromContext(ctx, "business.production")
	log.Info().Msg("starting daily production cycle")

	businesses, err := p.queries.GetActiveBusinesses(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active businesses: %w", err)
	}

	successCount := 0
	failCount := 0

	for _, biz := range businesses {
		err := p.processBusiness(ctx, biz)
		if err != nil {
			log.Warn().Err(err).Str("business_id", biz.ID).Msg("business production failed")
			failCount++
		} else {
			successCount++
		}
	}

	log.Info().
		Int("total_businesses", len(businesses)).
		Int("successful", successCount).
		Int("failed", failCount).
		Msg("daily production cycle completed")

	return nil
}

func (p *ProductionManager) processBusiness(ctx context.Context, biz sqlc.Business) error {
	log := logger.FromContext(ctx, "business.production")

	// 1. Parse production config
	var config ProductionConfig
	if len(biz.ProductionConfig) > 0 {
		if err := json.Unmarshal(biz.ProductionConfig, &config); err != nil {
			return fmt.Errorf("failed to parse production config: %w", err)
		}
	} else {
		return fmt.Errorf("missing production config")
	}

	// 2. Parse current inventory
	var inv Inventory
	if len(biz.Inventory) > 0 {
		if err := json.Unmarshal(biz.Inventory, &inv); err != nil {
			return fmt.Errorf("failed to parse inventory: %w", err)
		}
	} else {
		inv = Inventory{Inputs: make(map[string]decimal.Decimal), Outputs: make(map[string]decimal.Decimal)}
	}

	// 3. Fetch logged labor hours for today
	laborRecords, err := p.queries.GetDailyLaborByBusiness(ctx, biz.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch daily labor: %w", err)
	}

	workerCount := len(laborRecords)
	if workerCount == 0 {
		log.Debug().Str("business_id", biz.ID).Msg("no workers logged, skipping production")
		return nil
	}

	// Cap workers to max capacity
	effectiveWorkers := decimal.NewFromInt(int64(workerCount))
	if effectiveWorkers.GreaterThan(decimal.NewFromInt(int64(config.MaxWorkers))) {
		effectiveWorkers = decimal.NewFromInt(int64(config.MaxWorkers))
	}

	// 4. Calculate required inputs and potential outputs
	requiredInputs := config.InputRequired.Mul(effectiveWorkers)
	potentialOutputs := config.OutputPerWorker.Mul(effectiveWorkers)

	// Service businesses don't consume physical inputs
	isService := config.InputItemType == "NONE" || config.InputRequired.IsZero()

	if !isService {
		currentInputs := inv.Inputs[config.InputItemType]
		if currentInputs.LessThan(requiredInputs) {
			// Produce proportionally based on available inputs
			if currentInputs.IsZero() {
				return fmt.Errorf("insufficient inputs (0/%s)", requiredInputs.String())
			}
			ratio := currentInputs.Div(requiredInputs)
			potentialOutputs = potentialOutputs.Mul(ratio)
			requiredInputs = currentInputs // Consume all available
		}
	}

	// 5. Update inventory
	if !isService {
		inv.Inputs[config.InputItemType] = inv.Inputs[config.InputItemType].Sub(requiredInputs)
		if inv.Inputs[config.InputItemType].LessThan(decimal.Zero) {
			inv.Inputs[config.InputItemType] = decimal.Zero
		}
	}

	inv.Outputs[config.OutputItemType] = inv.Outputs[config.OutputItemType].Add(potentialOutputs)

	// 6. Save updated inventory to DB
	invJSON, err := json.Marshal(inv)
	if err != nil {
		return fmt.Errorf("failed to marshal updated inventory: %w", err)
	}

	err = p.queries.UpdateInventory(ctx, sqlc.UpdateInventoryParams{
		ID:        biz.ID,
		Inventory: invJSON,
	})
	if err != nil {
		return fmt.Errorf("failed to update inventory in DB: %w", err)
	}

	log.Info().
		Str("business_id", biz.ID).
		Str("produced", potentialOutputs.String()).
		Str("item", config.OutputItemType).
		Msg("business production successful")

	return nil
}