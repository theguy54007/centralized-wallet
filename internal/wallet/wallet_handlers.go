package wallet

import (
	"centralized-wallet/internal/apperrors"
	"centralized-wallet/internal/transaction"
	"errors"

	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BalanceHandler returns the wallet balance of the authenticated user
func BalanceHandler(ws WalletServiceInterface) gin.HandlerFunc {
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

// DepositHandler handles deposit requests and returns the updated balance and updated_at time
func DepositHandler(ws WalletServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from the context (set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"status": "error",
				"error":  "User ID not found",
			})
			return
		}

		var request struct {
			Amount float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "Invalid request data",
			})
			return
		}

		// Perform the deposit and get the updated Wallet struct
		wallet, err := ws.Deposit(userID.(int), request.Amount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		// Return structured success response with balance and updated_at timestamp
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"message":    "Deposit successful",
				"balance":    wallet.Balance,
				"updated_at": wallet.UpdatedAt, // the time wallet was last updated (after deposit)
			},
		})
	}
}

// WithdrawHandler handles withdraw requests and returns the updated balance and updated_at time
func WithdrawHandler(ws WalletServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from the context (set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "User ID not found"})
			return
		}

		var request struct {
			Amount float64 `json:"amount"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid request data"})
			return
		}

		// Perform the withdrawal and get the updated Wallet struct
		wallet, err := ws.Withdraw(userID.(int), request.Amount)
		if err != nil {
			// Check if the error is due to insufficient funds and return 400 Bad Request
			if err.Error() == "insufficient funds" {
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Insufficient funds"})
				return
			}
			// Otherwise, return 500 Internal Server Error for other issues
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
			return
		}

		// Return structured success response with balance and updated_at timestamp
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"message":    "Withdrawal successful",
				"balance":    wallet.Balance,
				"updated_at": wallet.UpdatedAt,
			},
		})
	}
}

func TransferHandler(ws WalletServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		fromUserID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"status": "error", "error": "User ID not found"})
			return
		}

		// Parse and validate request payload
		var request struct {
			ToWalletNumber string  `json:"to_wallet_number" binding:"required"`
			Amount         float64 `json:"amount" binding:"required,gt=0"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Invalid request data"})
			return
		}

		// Perform the transfer and get the updated Wallet for the from_user
		wallet, err := ws.Transfer(fromUserID.(int), request.ToWalletNumber, request.Amount)
		if err != nil {
			// Handle specific error cases and return 400 for user-related or validation issues
			switch err.Error() {
			case "from_user_id does not exist":
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "from_user_id does not exist"})
				return
			case "to_user_id does not exist":
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "to_user_id does not exist"})
				return
			case "insufficient funds":
				c.JSON(http.StatusBadRequest, gin.H{"status": "error", "error": "Insufficient funds"})
				return
			default:
				// Any other errors will return 500 Internal Server Error
				c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
				return
			}
		}

		// If transfer is successful, return the updated wallet balance and timestamp
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data": gin.H{
				"message":    "Transfer successful",
				"balance":    wallet.Balance,
				"updated_at": wallet.UpdatedAt,
			},
		})
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

		_, ok := userIDInterface.(int)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			return
		}

		// Get the wallet number from the context (set by WalletNumberMiddleware)
		walletNumberInterface, exists := c.Get("wallet_number")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Wallet number not found"})
			return
		}

		walletNumber, ok := walletNumberInterface.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid wallet number"})
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

		// Get the transaction history using the wallet number from the service
		transactions, err := ts.GetTransactionHistory(walletNumber, orderBy, limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Return the transactions as JSON
		c.JSON(http.StatusOK, gin.H{"transactions": transactions})
	}
}

func CreateWalletHandler(ws WalletServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from the context (set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
			return
		}

		wallet, err := ws.CreateWallet(userID.(int))
		if err != nil {
			if errors.Is(err, apperrors.ErrWalletAlreadyExists) {
				// If the user already has a wallet, return a 409 Conflict status
				c.JSON(http.StatusConflict, gin.H{"error": "User already has a wallet"})
			} else {
				// For any other error, return a 500 Internal Server Error
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		// Return structured success response with balance and updated_at timestamp
		c.JSON(http.StatusOK, gin.H{"wallet_number": wallet.WalletNumber})
	}
}
