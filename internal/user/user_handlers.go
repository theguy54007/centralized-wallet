package user

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"centralized-wallet/internal/auth"
	"centralized-wallet/internal/models"
)

// HTTP handler for user registration
func RegistrationHandler(us UserServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=6"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		user, err := us.RegisterUser(request.Email, request.Password)
		if err != nil {
			if err.Error() == "email already in use" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Email already in use"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User registered successfully",
			"user":    models.User{ID: user.ID, Email: user.Email},
		})
	}
}

// LoginHandler handles user login requests
func LoginHandler(us UserServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Authenticate the user
		user, err := us.LoginUser(request.Email, request.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// Generate a JWT token
		token, err := auth.GenerateJWT(user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Login successful",
			"token":   token,
			"user":    user,
		})
	}
}

func LogoutHandler(blacklistService *auth.BlacklistService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token is required"})
			return
		}

		// Ensure the token has "Bearer" prefix and remove it
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			return
		}

		// Validate the JWT token (no need to check claims, just validate structure)
		token, err := auth.ValidateJWT(tokenString)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Add the token to the blacklist
		err = blacklistService.BlacklistToken(tokenString, token)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to blacklist token"})
			return
		}

		// Return success message
		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}
