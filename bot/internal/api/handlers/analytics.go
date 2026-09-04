package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/0xProgress/simlife/bot/db/sqlc"
	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/rs/zerolog"
)

// AnalyticsHandler handles internal API endpoints used by the Python analytics service.
type AnalyticsHandler struct {
	queries *sqlc.Queries
	log     zerolog.Logger
}

// NewAnalyticsHandler initializes the analytics handler.
func NewAnalyticsHandler(queries *sqlc.Queries, log zerolog.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{
		queries: queries,
		log:     log.With().Str("handler", "analytics").Logger(),
	}
}

// AnalyticsSnapshotRequest is the payload posted by the Python analytics service.
type AnalyticsSnapshotRequest struct {
	EconomicDay int32           `json:"economic_day"`
	Metrics     json.RawMessage `json:"metrics"` // JSONB blob of computed metrics
}

// AnalyticsAlertRequest is the payload for anomaly alerts from the analytics service.
type AnalyticsAlertRequest struct {
	FlagType          string   `json:"flag_type"`
	ImplicatedPlayers []string `json:"implicated_players"`
	EvidenceSummary   string   `json:"evidence_summary"`
}

// HandleAnalyticsSnapshot receives the daily economic snapshot from the Python analytics service.
// This endpoint is authenticated by a shared secret (X-Analytics-Secret header), not a player JWT.
func (h *AnalyticsHandler) HandleAnalyticsSnapshot(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.analytics")

	var req AnalyticsSnapshotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(fmt.Errorf("failed to decode snapshot: %w", err)).Msg("invalid snapshot payload")
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.EconomicDay <= 0 {
		writeError(w, http.StatusBadRequest, "economic_day must be positive")
		return
	}

	// Store the snapshot in the database
	_, err := h.queries.RecordEconomicSnapshot(r.Context(), sqlc.RecordEconomicSnapshotParams{
		EconomicDay: req.EconomicDay,
		Metrics:     req.Metrics,
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to record snapshot: %w", err)).
			Int32("economic_day", req.EconomicDay).
			Msg("snapshot recording failed")
		writeError(w, http.StatusInternalServerError, "failed to record snapshot")
		return
	}

	log.Info().
		Int32("economic_day", req.EconomicDay).
		Msg("analytics snapshot recorded successfully")

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "recorded",
	})
}

// HandleAnalyticsAlert receives anomaly detection alerts from the Python analytics service.
func (h *AnalyticsHandler) HandleAnalyticsAlert(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.analytics")

	var req AnalyticsAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(fmt.Errorf("failed to decode alert: %w", err)).Msg("invalid alert payload")
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.FlagType == "" {
		writeError(w, http.StatusBadRequest, "flag_type is required")
		return
	}

	// Serialize implicated player IDs as JSON array
	playersJSON, err := json.Marshal(req.ImplicatedPlayers)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to serialize player IDs")
		return
	}

	// Store the anomaly flag
	_, err = h.queries.CreateAnomalyFlag(r.Context(), sqlc.CreateAnomalyFlagParams{
		FlagType:            req.FlagType,
		ImplicatedPlayerIds: playersJSON,
		EvidenceSummary:     req.EvidenceSummary,
	})
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to create anomaly flag: %w", err)).
			Str("flag_type", req.FlagType).
			Msg("anomaly flag creation failed")
		writeError(w, http.StatusInternalServerError, "failed to record alert")
		return
	}

	log.Warn().
		Str("flag_type", req.FlagType).
		Int("implicated_players", len(req.ImplicatedPlayers)).
		Msg("anomaly alert recorded")

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "recorded",
	})
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