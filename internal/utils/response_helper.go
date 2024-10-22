package utils

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, err *AppError, internalErr error, context string) {
	if internalErr != nil {
		logError(internalErr, err.Message, context)
	}
	c.JSON(err.Code, APIResponse{
		Status:  "error",
		Message: err.Message,
	})
}

func logError(err error, message string, context string) {
	// Log the error details safely
	log.Printf("[ERROR] %s: %s, Details: %v", context, message, err)
}
