package middleware

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
)

// ResponseTransformMiddleware is a SOTA middleware that transforms snake_case to camelCase
func ResponseTransformMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a custom writer
		w := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = w

		// Process the request
		c.Next()

		// Transform the response if it's JSON
		if w.Header().Get("Content-Type") == "application/json" ||
			strings.Contains(w.Header().Get("Content-Type"), "application/json") {
			// Parse the response
			var data interface{}
			if err := json.Unmarshal(w.body.Bytes(), &data); err == nil {
				// Check if this response needs transformation
				// Skip transformation for paths that might have pointer field issues
				path := c.Request.URL.Path
				if shouldSkipTransformation(path) {
					// Write original response without transformation
					w.ResponseWriter.Write(w.body.Bytes())
					return
				}

				// Transform to camelCase only for safe paths
				transformed := transformToCamelCase(data)

				// Write transformed response
				transformedBytes, _ := json.Marshal(transformed)
				// Content-Length will be set automatically by Gin
				w.ResponseWriter.Write(transformedBytes)
				return
			}
		}

		// Write original response if not JSON or transformation failed
		w.ResponseWriter.Write(w.body.Bytes())
	}
}

// shouldSkipTransformation determines if a path should skip the transformation
// to avoid pointer field issues
func shouldSkipTransformation(path string) bool {
	// Skip transformation for museum APIs that have pointer fields
	skipPaths := []string{
		"/api/v1/museum/entries",
		"/api/v1/museum/entries/",
	}
	
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	
	return false
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	// Capture the response body
	w.body.Write(b)
	// Don't write to the actual response yet
	return len(b), nil
}

func (w *responseWriter) WriteString(s string) (int, error) {
	// Capture string responses too
	return w.Write([]byte(s))
}

// transformToCamelCase recursively transforms snake_case keys to camelCase
func transformToCamelCase(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for key, value := range v {
			camelKey := snakeToCamelCase(key)
			newMap[camelKey] = transformToCamelCase(value)
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(v))
		for i, value := range v {
			newSlice[i] = transformToCamelCase(value)
		}
		return newSlice
	default:
		return data
	}
}

// snakeToCamelCase converts snake_case to camelCase
func snakeToCamelCase(s string) string {
	// Handle special cases
	if s == "id" || s == "ok" {
		return s
	}

	// Split by underscore
	parts := strings.Split(s, "_")
	if len(parts) == 1 {
		return s
	}

	// Convert to camelCase
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}

	return result
}
