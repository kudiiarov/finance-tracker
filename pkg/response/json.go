// Package response provides HTTP response utilities.
package response

import (
	"encoding/json"
	"net/http"
)

// JSON sends a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Error sends a JSON error response.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

// Success sends a JSON success message.
func Success(w http.ResponseWriter, message string) {
	JSON(w, http.StatusOK, map[string]string{"message": message})
}
