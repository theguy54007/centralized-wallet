package utils

import (
	// "centralized-wallet/internal/apperrors"

	"github.com/gin-gonic/gin"
)

func ErrorMiddlewareHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check if the error is an AppError, else return a generic server error
				if appErr, ok := err.(*AppError); ok {
					ErrorResponse(c, appErr)
				} else {
					ErrorResponse(c, ErrInternalServerError)
				}
				c.Abort()
			}
		}()
		c.Next()
	}
}
