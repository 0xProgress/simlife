package api

import (
	"encoding/json"
	"net/http"

	"github.com/0xProgress/simlife/bot/internal/logger"
)

// All route handler methods are defined on *Server in their respective
// handler files (activity.go, player.go, market.go, city.go, business.go, analytics.go).
//
// This file contains only shared HTTP response helpers used across all handlers.

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Headers already sent; log the error but can't change status
			logger.Package("api").Error().Err(err).Msg("failed to encode JSON response")
		}
	}
}

// writeError writes a standardized JSON error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// readJSON decodes the request body into the provided target struct.
// Returns an error if the body is malformed.
func readJSON(r *http.Request, target interface{}) error {
	if r.Body == nil {
		return http.ErrBodyNotAllowed
	}
	defer r.Body.Close()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Strict parsing — reject unexpected fields
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}