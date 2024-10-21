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

		statusCode := c.Writer.Status()

		// Log only if the response status code indicates an error (400 and above)
		if statusCode >= 400 {
			duration := time.Since(startTime)
			method := c.Request.Method
			path := c.Request.URL.Path
			clientIP := c.ClientIP()
			responseBody := responseWriter.body.String()

			// Log request and error response details
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
