// Package response provides shared HTTP response utilities
// Safe to import - doesn't affect existing implementations
package response

import (
	"encoding/json"
	"net/http"
)

// JSONResponse sends a JSON response with status code
func JSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// JSONSuccess sends a success JSON response
func JSONSuccess(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// JSONError sends an error JSON response
func JSONError(w http.ResponseWriter, code int, message string) {
	JSON(w, code, map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// JSONMessage sends a simple message response
func JSONMessage(w http.ResponseWriter, code int, message string) {
	JSON(w, code, map[string]interface{}{
		"success": true,
		"message": message,
	})
}