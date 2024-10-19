package wallet

import (
	"centralized-wallet/internal/transaction"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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

// DepositHandler allows the user to deposit money into their wallet
func DepositHandler(ws *WalletService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from the context (set by JWTMiddleware)
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
		// Get user ID from the context (set by JWTMiddleware)
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

		// Check if the from_user_id exists
		existsFromUser, err := ws.UserExists(fromUserID.(int))
		if err != nil || !existsFromUser {
			c.JSON(http.StatusBadRequest, gin.H{"error": "from_user_id does not exist"})
			return
		}

		// Check if the to_user_id exists
		existsToUser, err := ws.UserExists(request.ToUserID)
		if err != nil || !existsToUser {
			c.JSON(http.StatusBadRequest, gin.H{"error": "to_user_id does not exist"})
			return
		}

		// Perform the transfer
		err = ws.Transfer(fromUserID.(int), request.ToUserID, request.Amount)
		if err != nil {
			// Handle other errors
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// If transfer is successful
		c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})
	}
}

// TransactionHistoryHandler returns the transaction history for the authenticated user
func TransactionHistoryHandler(ts transaction.TransactionServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user ID from the context (set by JWT middleware)
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
			return
		}

		userID, ok := userIDInterface.(int)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			return
		}

		// Parse query parameters for sorting and limiting
		orderBy := c.DefaultQuery("order", "desc")
		if orderBy != "asc" && orderBy != "desc" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order, must be 'asc' or 'desc'"})
			return
		}

		const maxLimit = 100
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > maxLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit, must be between 1 and 100"})
			return
		}

		// Get the transaction history from the service
		transactions, err := ts.GetTransactionHistory(userID, orderBy, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return the transactions as JSON
		c.JSON(http.StatusOK, gin.H{"transactions": transactions})
	}
}
