package logging

import (
	"bytes"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseWriter is a custom writer to capture response body
type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w ResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggingMiddlewareForErrors logs request details and response only for errors (status 400 and above)
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Create a custom response writer to capture the response
		responseWriter := &ResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// After request processing
		statusCode := c.Writer.Status()
		duration := time.Since(startTime)
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()
		responseBody := responseWriter.body.String()

		// Get the internal error from the context if present
		internalError, exists := c.Get("internal_error")
		if exists {
			// Log the internal error (if present) along with the request details
			Log.WithFields(map[string]interface{}{
				"status":         statusCode,
				"method":         method,
				"path":           path,
				"client_ip":      clientIP,
				"duration":       duration,
				"response":       responseBody,
				"internal_error": internalError,
			}).Error("Internal error response logged")
		} else if statusCode >= 400 {
			// Log error response details (for client errors without an internal error)
			Log.WithFields(map[string]interface{}{
				"status":    statusCode,
				"method":    method,
				"path":      path,
				"client_ip": clientIP,
				"duration":  duration,
				"response":  responseBody,
			}).Error("Error response logged")
		}
	}
}
