package handlers

import (
	"fmt"
	"net/http"

	"github.com/0xProgress/simlife/bot/internal/logger"
	"github.com/0xProgress/simlife/bot/internal/world"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// CityHandler handles city state API endpoints for the Activity's City View.
type CityHandler struct {
	city *world.CityManager
	log  zerolog.Logger
}

// NewCityHandler initializes the city handler.
func NewCityHandler(city *world.CityManager, log zerolog.Logger) *CityHandler {
	return &CityHandler{
		city: city,
		log:  log.With().Str("handler", "city").Logger(),
	}
}

// HandleGetCity returns the full city state (all plots, businesses, online players).
// This endpoint serves cached Redis data when available, only hitting PostgreSQL on cache miss.
func (h *CityHandler) HandleGetCity(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.city")

	state, err := h.city.GetCityState(r.Context())
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch city state: %w", err)).Msg("city state fetch failed")
		writeError(w, http.StatusInternalServerError, "failed to fetch city state")
		return
	}

	writeJSON(w, http.StatusOK, state)
}

// HandleGetDistrict returns aggregated statistics for a specific zone/district.
func (h *CityHandler) HandleGetDistrict(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context(), "handlers.city")

	zoneStr := chi.URLParam(r, "zone")
	if zoneStr == "" {
		writeError(w, http.StatusBadRequest, "zone parameter is required")
		return
	}

	zoneType := world.ZoneType(zoneStr)
	if !world.IsValidZoneType(zoneStr) {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("invalid zone type: %s", zoneStr))
		return
	}

	summary, err := h.city.GetDistrictSummary(r.Context(), zoneType)
	if err != nil {
		log.Error().Err(fmt.Errorf("failed to fetch district summary: %w", err)).
			Str("zone", zoneStr).
			Msg("district summary fetch failed")
		writeError(w, http.StatusInternalServerError, "failed to fetch district summary")
		return
	}

	writeJSON(w, http.StatusOK, summary)
}