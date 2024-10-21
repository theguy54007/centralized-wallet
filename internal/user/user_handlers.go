package user

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"centralized-wallet/internal/auth"
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/utils"
)

// HTTP handler for user registration
func RegistrationHandler(us UserServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}

		// Validate the request body
		if err := c.ShouldBindJSON(&request); err != nil {
			if err.Error() == "Key: 'Email' Error:Field validation for 'Email' failed" {
				utils.ErrorResponse(c, utils.ErrInvalidEmailFormat, nil, "")
				return
			}
			if err.Error() == "Key: 'Password' Error:Field validation for 'Password' failed" {
				utils.ErrorResponse(c, utils.ErrPasswordTooShort, nil, "")
				return
			}
			utils.ErrorResponse(c, utils.ErrInvalidRequest, nil, "")
			return
		}

		// Register the user
		user, err := us.RegisterUser(request.Email, request.Password)
		if err != nil {
			if errors.Is(err, utils.ErrEmailAlreadyInUse) {
				utils.ErrorResponse(c, utils.ErrEmailAlreadyInUse, nil, "")
				return
			}
			utils.ErrorResponse(c, utils.ErrUserCreationFailed, err, "")
			return
		}

		// Success response
		utils.SuccessResponse(c, utils.MsgUserRegistered, models.User{ID: user.ID, Email: user.Email})
	}
}

// LoginHandler handles user login requests
func LoginHandler(us UserServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		// Validate the input
		if err := c.ShouldBindJSON(&request); err != nil {
			utils.ErrorResponse(c, utils.ErrInvalidRequest, nil, "")
			return
		}

		// Authenticate the user
		user, err := us.LoginUser(request.Email, request.Password)
		if err != nil {
			utils.ErrorResponse(c, utils.ErrInvalidCredentials, nil, "")
			return
		}

		// Generate a JWT token
		token, err := auth.GenerateJWT(user.ID)
		if err != nil {
			utils.ErrorResponse(c, utils.ErrTokenGenerationFailed, nil, "")
			return
		}

		// Success response with token
		utils.SuccessResponse(c, "Login successful", gin.H{
			"token": token,
			"user":  user,
		})
	}
}

func LogoutHandler(blacklistService *auth.BlacklistService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get the token string from context (set by JWT middleware)
		tokenString, exists := c.Get("token_string")
		if !exists {
			utils.ErrorResponse(c, utils.ErrUnauthorized, nil, "")
			return
		}

		// Get the token from context (set by JWT middleware)
		token, exists := c.Get("token")
		if !exists || token == nil {
			utils.ErrorResponse(c, utils.ErrInvalidToken, nil, "")
			return
		}

		// Add the token to the blacklist
		err := blacklistService.BlacklistToken(tokenString.(string), token.(*jwt.Token))
		if err != nil {
			utils.ErrorResponse(c, utils.ErrInternalServerError, err, "")
			return
		}

		// Return success message
		utils.SuccessResponse(c, "Logged out successfully", nil)
	}
}
