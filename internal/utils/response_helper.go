package utils

import (
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
		// Store internal error in the context to be logged by the middleware
		c.Set("internal_error", internalErr.Error())
	}

	c.JSON(err.Code, APIResponse{
		Status:  "error",
		Message: err.Message,
	})
}
