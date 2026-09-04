package business

import "github.com/shopspring/decimal"

// ProductionConfig defines the recipe and capacity for a business.
// Stored as JSONB in the `businesses` table.
type ProductionConfig struct {
	InputItemType   string          `json:"input_item_type"`
	InputRequired   decimal.Decimal `json:"input_required"`   // Inputs consumed per worker per day
	OutputItemType  string          `json:"output_item_type"`
	OutputPerWorker decimal.Decimal `json:"output_per_worker"` // Outputs produced per worker per day
	MaxWorkers      int             `json:"max_workers"`
}

// Inventory tracks the current stock of a business.
// Stored as JSONB in the `businesses` table.
type Inventory struct {
	Inputs  map[string]decimal.Decimal `json:"inputs"`  // item_type -> quantity
	Outputs map[string]decimal.Decimal `json:"outputs"` // item_type -> quantity
}