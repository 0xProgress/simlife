package business

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/economy"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

var (
	ErrInsufficientCapital = errors.New("insufficient capital to open business")
	ErrBusinessNotFound    = errors.New("business not found")
	ErrUnauthorized        = errors.New("unauthorized to modify business")
)

// Engine coordinates all business operations: opening, upgrading, and closing.
type Engine struct {
	queries *sqlc.Queries
	ledger  *economy.Ledger
	log     zerolog.Logger
}

// NewEngine initializes the business engine.
func NewEngine(q *sqlc.Queries, l *economy.Ledger, log zerolog.Logger) *Engine {
	return &Engine{
		queries: q,
		ledger:  l,
		log:     log.With().Str("component", "business_engine").Logger(),
	}
}

// OpenBusiness validates capital, charges the registration fee, and creates the business record.
func (e *Engine) OpenBusiness(ctx context.Context, ownerID, businessType, name, cityPlotID string) (string, error) {
	log := logger.FromContext(ctx, "business.engine")

	// 1. Validate business type and get registration fee
	regFee := getRegistrationFee(businessType)
	if regFee.IsZero() {
		return "", fmt.Errorf("invalid or unsupported business type: %s", businessType)
	}

	// 2. Fetch owner's wallet and treasury
	ownerWallet, err := e.queries.GetAccountByType(ctx, sqlc.GetAccountByTypeParams{
		PlayerID:    ownerID,
		AccountType: sqlc.AccountTypeWALLET,
	})
	if err != nil {
		return "", fmt.Errorf("failed to fetch owner wallet: %w", err)
	}

	treasury, err := e.queries.GetTreasuryAccount(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch treasury: %w", err)
	}

	// 3. Charge registration fee via Ledger (Atomic Double-Entry)
	err = e.ledger.Transfer(ctx, ownerWallet.ID, treasury.ID, regFee, "BUSINESS_REGISTRATION", ownerID, fmt.Sprintf("Business registration: %s", name))
	if err != nil {
		if errors.Is(err, economy.ErrInsufficientFunds) {
			return "", ErrInsufficientCapital
		}
		return "", fmt.Errorf("failed to charge registration fee: %w", err)
	}

	// 4. Initialize default inventory and production config
	defaultConfig := getDefaultProductionConfig(businessType)
	configJSON, err := json.Marshal(defaultConfig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal production config: %w", err)
	}

	defaultInventory := Inventory{
		Inputs:  make(map[string]decimal.Decimal),
		Outputs: make(map[string]decimal.Decimal),
	}
	invJSON, err := json.Marshal(defaultInventory)
	if err != nil {
		return "", fmt.Errorf("failed to marshal inventory: %w", err)
	}

	// 5. Create DB record
	biz, err := e.queries.CreateBusiness(ctx, sqlc.CreateBusinessParams{
		OwnerID:          ownerID,
		BusinessType:     businessType,
		Name:             name,
		CityPlotID:       pgtype.Text{String: cityPlotID, Valid: cityPlotID != ""},
		Inventory:        invJSON,
		ProductionConfig: configJSON,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create business record: %w", err)
	}

	log.Info().
		Str("owner_id", ownerID).
		Str("business_id", biz.ID).
		Str("type", businessType).
		Str("fee", regFee.String()).
		Msg("business opened successfully")

	return biz.ID, nil
}

// CloseBusiness validates ownership and marks the business as closed.
func (e *Engine) CloseBusiness(ctx context.Context, businessID, ownerID string) error {
	log := logger.FromContext(ctx, "business.engine")

	biz, err := e.queries.GetBusinessByID(ctx, businessID)
	if err != nil {
		return ErrBusinessNotFound
	}

	if biz.OwnerID != ownerID {
		return ErrUnauthorized
	}

	err = e.queries.UpdateBusinessStatus(ctx, sqlc.UpdateBusinessStatusParams{
		ID:     businessID,
		Status: "CLOSED",
	})
	if err != nil {
		return fmt.Errorf("failed to update business status: %w", err)
	}

	log.Info().Str("business_id", businessID).Msg("business closed")
	return nil
}

func getRegistrationFee(businessType string) decimal.Decimal {
	switch businessType {
	case "BAKERY", "FORGE", "CLINIC", "SERVICE":
		return decimal.NewFromInt(5000)
	case "PRODUCTION", "MANUFACTURING":
		return decimal.NewFromInt(15000)
	default:
		return decimal.Zero
	}
}

func getDefaultProductionConfig(businessType string) ProductionConfig {
	switch businessType {
	case "BAKERY":
		return ProductionConfig{
			InputItemType:   "WHEAT",
			InputRequired:   decimal.NewFromInt(2),
			OutputItemType:  "BREAD",
			OutputPerWorker: decimal.NewFromInt(10),
			MaxWorkers:      5,
		}
	case "FORGE":
		return ProductionConfig{
			InputItemType:   "IRON_ORE",
			InputRequired:   decimal.NewFromInt(3),
			OutputItemType:  "TOOLS",
			OutputPerWorker: decimal.NewFromInt(5),
			MaxWorkers:      5,
		}
	default:
		// Service businesses do not consume physical inputs
		return ProductionConfig{
			InputItemType:   "NONE",
			InputRequired:   decimal.Zero,
			OutputItemType:  "SERVICE",
			OutputPerWorker: decimal.NewFromInt(1),
			MaxWorkers:      10,
		}
	}
}