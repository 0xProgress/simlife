package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
)

// BusinessHandler handles business dashboard API endpoints.
type BusinessHandler struct {
	queries *sqlc.Queries
	log     zerolog.Logger
}

// NewBusinessHandler initializes the business handler.
func NewBusinessHandler(queries *sqlc.Queries, log zerolog.Logger) *BusinessHandler {
	return &BusinessHandler{
		queries: queries,
		log:     log.With().Str("handler", "business").Logger(),
	}
}

// BusinessResponse represents a business in API responses.
type BusinessResponse struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Type             string          `json:"type"`
	Status           string          `json:"status"`
	OwnerID          string          `json:"owner_id"`
	EmployeeCount    int             `json:"employee_count"`
	Inventory        json.RawMessage `json:"inventory"`
	ProductionConfig json.RawMessage `json:"production_config"`
	CreatedAt        string          `json:"created_at"`
}

// WorkerResponse represents an employee in API responses.
type WorkerResponse struct {
	EmployeeID   string `json:"employee_id"`
	Username     string `json:"username"`
	WageRate     string `json:"wage_rate"`
	MinDailyHours string `json:"min_daily_hours"`
	Status       string `json:"status"`
	StartDate    string `json:"start_date"`
}

// FinancialSummaryResponse represents a business's financial summary.
type FinancialSummaryResponse struct {
	BusinessID       string `json:"business_id"`
	ProjectedRevenue string `json:"projected_revenue"`
	ProjectedCost    string `json:"projected_cost"`
	NetProfit        string `json:"net_profit"`
	EmployeeCount    int    `json:"employee_count"`
}

// HandleGetMyBusinesses returns all businesses owned by the authenticated player.
func (h *BusinessHandler) HandleGetMyBusinesses(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.business")

	playerID := getPlayerIDFromContext(r.Context())
	if playerID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	businesses, err := h.queries.GetBusinessesByOwner(r.Context(), playerID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch businesses: %w", err)).Msg("businesses fetch failed")
		writeError(w, http.StatusInternalServerError, "failed to fetch businesses")
		return
	}

	responses := make([]BusinessResponse, 0, len(businesses))
	for _, biz := range businesses {
		// Count employees
		employees, _ := h.queries.GetEmploymentByBusiness(r.Context(), biz.ID)
		empCount := len(employees)

		responses = append(responses, BusinessResponse{
			ID:               biz.ID,
			Name:             biz.Name,
			Type:             biz.BusinessType,
			Status:           string(biz.Status),
			OwnerID:          biz.OwnerID,
			EmployeeCount:    empCount,
			Inventory:        biz.Inventory,
			ProductionConfig: biz.ProductionConfig,
			CreatedAt:        biz.CreatedAt.Time.String(),
		})
	}

	writeJSON(w, http.StatusOK, responses)
}

// HandleGetBusiness returns details for a specific business.
func (h *BusinessHandler) HandleGetBusiness(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.business")

	businessID := chi.URLParam(r, "business_id")
	if businessID == "" {
		writeError(w, http.StatusBadRequest, "business_id is required")
		return
	}

	biz, err := h.queries.GetBusinessByID(r.Context(), businessID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch business: %w", err)).Msg("business fetch failed")
		writeError(w, http.StatusNotFound, "business not found")
		return
	}

	employees, _ := h.queries.GetEmploymentByBusiness(r.Context(), biz.ID)

	writeJSON(w, http.StatusOK, BusinessResponse{
		ID:               biz.ID,
		Name:             biz.Name,
		Type:             biz.BusinessType,
		Status:           string(biz.Status),
		OwnerID:          biz.OwnerID,
		EmployeeCount:    len(employees),
		Inventory:        biz.Inventory,
		ProductionConfig: biz.ProductionConfig,
		CreatedAt:        biz.CreatedAt.Time.String(),
	})
}

// HandleGetBusinessWorkers returns all active employees for a business.
func (h *BusinessHandler) HandleGetBusinessWorkers(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.business")

	businessID := chi.URLParam(r, "business_id")
	if businessID == "" {
		writeError(w, http.StatusBadRequest, "business_id is required")
		return
	}

	employees, err := h.queries.GetEmploymentByBusiness(r.Context(), businessID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch employees: %w", err)).Msg("employees fetch failed")
		writeError(w, http.StatusInternalServerError, "failed to fetch employees")
		return
	}

	responses := make([]WorkerResponse, 0, len(employees))
	for _, emp := range employees {
		player, _ := h.queries.GetPlayerByID(r.Context(), emp.EmployeeID)
		username := "Unknown"
		if player != nil {
			username = player.Username
		}

		responses = append(responses, WorkerResponse{
			EmployeeID:    emp.EmployeeID,
			Username:      username,
			WageRate:      numericToDecimal(emp.WageRate).String(),
			MinDailyHours: numericToDecimal(emp.MinDailyHours).String(),
			Status:        string(emp.Status),
			StartDate:     emp.StartDate.Time.String(),
		})
	}

	writeJSON(w, http.StatusOK, responses)
}

// HandleGetBusinessFinancials returns a financial summary for a business.
func (h *BusinessHandler) HandleGetBusinessFinancials(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.business")

	businessID := chi.URLParam(r, "business_id")
	if businessID == "" {
		writeError(w, http.StatusBadRequest, "business_id is required")
		return
	}

	// Fetch employees to calculate projected payroll
	employees, err := h.queries.GetEmploymentByBusiness(r.Context(), businessID)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch employees: %w", err)).Msg("financials fetch failed")
		writeError(w, http.StatusInternalServerError, "failed to fetch financials")
		return
	}

	var totalWages decimal.Decimal
	for _, emp := range employees {
		wage := numericToDecimal(emp.WageRate)
		totalWages = totalWages.Add(wage)
	}

	// TODO: Calculate projected revenue from production config and market prices
	projectedRevenue := decimal.Zero
	projectedCost := totalWages // Simplified: cost = wages only for now
	netProfit := projectedRevenue.Sub(projectedCost)

	writeJSON(w, http.StatusOK, FinancialSummaryResponse{
		BusinessID:       businessID,
		ProjectedRevenue: projectedRevenue.String(),
		ProjectedCost:    projectedCost.String(),
		NetProfit:        netProfit.String(),
		EmployeeCount:    len(employees),
	})
}

// numericToDecimal converts pgtype.Numeric to shopspring/decimal.
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

// getPlayerIDFromContext extracts the player ID from the request context.
func getPlayerIDFromContext(ctx context.Context) string {
	type contextKey string
	if v, ok := ctx.Value(contextKey("api_player_id")).(string); ok {
		return v
	}
	return ""
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// writeError writes a standardized error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}