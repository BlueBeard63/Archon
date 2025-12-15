package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/BlueBeard63/archon-node/internal/models"
)

// AuthMiddleware checks for valid API key
func AuthMiddleware(apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get Authorization header
			auth := r.Header.Get("Authorization")
			if auth == "" {
				respondError(w, http.StatusUnauthorized, "Missing Authorization header")
				return
			}

			// Check for Bearer token
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, http.StatusUnauthorized, "Invalid Authorization header format")
				return
			}

			// Validate API key
			if parts[1] != apiKey {
				respondError(w, http.StatusUnauthorized, "Invalid API key")
				return
			}

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log request
		// fmt.Printf("[%s] %s %s\n", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// Helper functions for responses

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, models.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
	})
}
