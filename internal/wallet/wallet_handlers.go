package wallet

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// BalanceHandler returns the wallet balance of the user
func BalanceHandler(ws *WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDParam := c.Param("user_id")
		userID, err := strconv.Atoi(userIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		balance, err := ws.GetBalance(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"balance": balance})
	}
}

// DepositHandler allows the user to deposit money into their wallet
func DepositHandler(ws *WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDParam := c.Param("user_id")
		userID, err := strconv.Atoi(userIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var request struct {
			Amount float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		err = ws.Deposit(userID, request.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Deposit successful"})
	}
}

// WithdrawHandler allows the user to withdraw money from their wallet
func WithdrawHandler(ws *WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDParam := c.Param("user_id")
		userID, err := strconv.Atoi(userIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		var request struct {
			Amount float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		err = ws.Withdraw(userID, request.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Withdrawal successful"})
	}
}

// TransferHandler allows the user to transfer money to another user's wallet
func TransferHandler(ws *WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			FromUserID int     `json:"from_user_id"`
			ToUserID   int     `json:"to_user_id"`
			Amount     float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
			return
		}

		// Perform the transfer
		err := ws.Transfer(request.FromUserID, request.ToUserID, request.Amount)
		if err != nil {
			// Check for user existence errors
			if err.Error() == "to_user_id does not exist" || err.Error() == "from_user_id does not exist" {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
	}
}
