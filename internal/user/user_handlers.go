package user

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"centralized-wallet/internal/models"
)

// HTTP handler for user registration
func RegistrationHandler(us *UserService) gin.HandlerFunc {
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
func LoginHandler(us *UserService) gin.HandlerFunc {
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
		token, err := GenerateJWT(user.ID)
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
