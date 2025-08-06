package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

// responseTransformWriter captures response body for transformation
type responseTransformWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseTransformWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseTransformWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

// APITransformMiddleware transforms API responses from snake_case to camelCase
// This is a SOTA implementation with minimal overhead
func APITransformMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip transformation for non-JSON responses
		acceptHeader := c.GetHeader("Accept")
		if acceptHeader != "" && !strings.Contains(acceptHeader, "application/json") && !strings.Contains(acceptHeader, "*/*") {
			c.Next()
			return
		}

		// Capture the response
		rtw := &responseTransformWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = rtw

		// Process request
		c.Next()

		// Only transform successful JSON responses
		if rtw.Status() >= 200 && rtw.Status() < 300 {
			contentType := rtw.Header().Get("Content-Type")
			if strings.Contains(contentType, "application/json") && rtw.body.Len() > 0 {
				// Parse response
				var data interface{}
				if err := json.Unmarshal(rtw.body.Bytes(), &data); err == nil {
					// Transform keys
					transformed := transformKeys(data, snakeToCamel)
					
					// Marshal transformed data
					transformedBytes, _ := json.Marshal(transformed)
					
					// Write directly to original writer
					rtw.ResponseWriter.Header().Set("Content-Type", "application/json")
					rtw.ResponseWriter.WriteHeader(rtw.Status())
					rtw.ResponseWriter.Write(transformedBytes)
					return
				}
			}
		}
		
		// If no transformation needed, write original response
		rtw.ResponseWriter.WriteHeader(rtw.Status())
		rtw.ResponseWriter.Write(rtw.body.Bytes())
	}
}

// transformKeys recursively transforms object keys
func transformKeys(data interface{}, transform func(string) string) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		newMap := make(map[string]interface{})
		for key, value := range v {
			newKey := transform(key)
			newMap[newKey] = transformKeys(value, transform)
		}
		return newMap
	case []interface{}:
		newSlice := make([]interface{}, len(v))
		for i, value := range v {
			newSlice[i] = transformKeys(value, transform)
		}
		return newSlice
	default:
		return data
	}
}

// snakeToCamel converts snake_case to camelCase
func snakeToCamel(s string) string {
	// Special cases for common fields
	switch s {
		case "id", "ok":
			return s
		case "created_at":
			return "createdAt"
		case "updated_at":
			return "updatedAt"
		case "deleted_at":
			return "deletedAt"
	}
	
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// camelToSnake converts camelCase to snake_case (for request transformation)
func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// RequestTransformMiddleware transforms incoming requests from camelCase to snake_case
func RequestTransformMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only transform JSON requests
		if c.ContentType() != "application/json" || c.Request.ContentLength == 0 {
			c.Next()
			return
		}

		// Read body
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Parse JSON
		var data interface{}
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			c.Next()
			return
		}

		// Transform keys
		transformed := transformKeys(data, camelToSnake)
		
		// Write transformed body back
		transformedBytes, err := json.Marshal(transformed)
		if err != nil {
			c.Next()
			return
		}
		
		c.Request.Body = io.NopCloser(bytes.NewBuffer(transformedBytes))
		c.Request.ContentLength = int64(len(transformedBytes))
		
		c.Next()
	}
}