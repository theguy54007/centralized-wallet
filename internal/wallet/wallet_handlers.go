package wallet

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// BalanceHandler returns the wallet balance of the authenticated user
func BalanceHandler(ws *WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
			return
		}

		// Fetch balance from the WalletService
		balance, err := ws.GetBalance(userID.(int))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"balance": balance})
	}
}

// DepositHandler allows the authenticated user to deposit money into their wallet
func DepositHandler(ws *WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
			return
		}

		var request struct {
			Amount float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Perform the deposit
		err := ws.Deposit(userID.(int), request.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Deposit successful"})
	}
}

// WithdrawHandler allows the authenticated user to withdraw money from their wallet
func WithdrawHandler(ws *WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
			return
		}

		var request struct {
			Amount float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Perform the withdrawal
		err := ws.Withdraw(userID.(int), request.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful"})
	}
}

func TransferHandler(ws *WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by JWTMiddleware)
		fromUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
			return
		}

		var request struct {
			ToUserID int     `json:"to_user_id"`
			Amount   float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Perform the transfer, check for user existence in the walletRepo.Transfer logic
		err := ws.Transfer(fromUserID.(int), request.ToUserID, request.Amount)
		if err != nil {
			// Handle specific errors related to user existence
			if err.Error() == "from_user_id does not exist" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "from_user_id does not exist"})
				return
			}
			if err.Error() == "to_user_id does not exist" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "to_user_id does not exist"})
				return
			}
			// Return a generic internal server error for other cases
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// If transfer is successful
		c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
	}
}
