package wallet

import (
	"centralized-wallet/internal/transaction"
	"centralized-wallet/internal/utils"

	"strconv"

	"github.com/gin-gonic/gin"
)

// BalanceHandler returns the wallet balance of the authenticated user
func BalanceHandler(ws WalletServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (already set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			// This should not happen because the middleware would have already handled it
			utils.ErrorResponse(c, utils.ErrUnauthorized)
			return
		}

		// Fetch balance from the WalletService
		balance, err := ws.GetBalance(userID.(int))
		if err != nil {
			// Handle specific error cases
			switch err {
			case utils.RepoErrWalletNotFound:
				utils.ErrorResponse(c, utils.ErrWalletNotFound)
				return
			default:
				utils.ErrorResponse(c, utils.ErrInternalServerError)
				return
			}
		}

		// Respond with balance
		utils.SuccessResponse(c, utils.MsgBalanceRetrieved, gin.H{"balance": balance})
	}
}

// DepositHandler handles deposit requests and returns the updated balance and updated_at time
func DepositHandler(ws WalletServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from the context (set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, utils.ErrUnauthorized)
			return
		}

		// Parse request body
		var request struct {
			Amount float64 `json:"amount" binding:"required,gt=0"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			utils.ErrorResponse(c, utils.ErrInvalidRequest)
			return
		}

		// Perform the deposit and get the updated Wallet struct
		wallet, err := ws.Deposit(userID.(int), request.Amount)
		if err != nil {
			switch err {
			case utils.RepoErrWalletNotFound:
				utils.ErrorResponse(c, utils.ErrWalletNotFound)
				return
			case utils.ErrDatabaseError:
				utils.ErrorResponse(c, utils.ErrDatabaseError)
				return
			default:
				utils.ErrorResponse(c, utils.ErrInternalServerError)
				return
			}
		}

		// Return structured success response
		utils.SuccessResponse(c, utils.MsgDepositSuccessful, gin.H{
			"balance":    wallet.Balance,
			"updated_at": wallet.UpdatedAt, // timestamp of last wallet update (after deposit)
		})
	}
}

// WithdrawHandler handles withdraw requests and returns the updated balance and updated_at time
func WithdrawHandler(ws WalletServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from the context (set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, utils.ErrUnauthorized)
			return
		}

		// Parse the request body
		var request struct {
			Amount float64 `json:"amount" binding:"required,gt=0"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			utils.ErrorResponse(c, utils.ErrInvalidRequest)
			return
		}

		// Perform the withdrawal and get the updated Wallet struct
		wallet, err := ws.Withdraw(userID.(int), request.Amount)
		if err != nil {
			switch err {
			case utils.RepoErrUserNotFound:
				utils.ErrorResponse(c, utils.ErrUserNotFound)
				return
			case utils.RepoErrInsufficientFunds:
				utils.ErrorResponse(c, utils.ErrorInsufficientFunds)
				return
			case utils.ErrDatabaseError:
				utils.ErrorResponse(c, utils.ErrDatabaseError)
			default:
				utils.ErrorResponse(c, utils.ErrInternalServerError)
				return
			}
		}

		// Return structured success response
		utils.SuccessResponse(c, utils.MsgWithdrawSuccessful, gin.H{
			"balance":    wallet.Balance,
			"updated_at": wallet.UpdatedAt, // last updated time
		})
	}
}

func TransferHandler(ws WalletServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the fromUserID from context (set by JWTMiddleware)
		fromUserID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, utils.ErrUnauthorized)
			return
		}

		// Parse and validate request payload
		var request struct {
			ToWalletNumber string  `json:"to_wallet_number" binding:"required"`
			Amount         float64 `json:"amount" binding:"required,gt=0"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			utils.ErrorResponse(c, utils.ErrInvalidRequest)
			return
		}

		// Perform the transfer operation
		wallet, err := ws.Transfer(fromUserID.(int), request.ToWalletNumber, request.Amount)
		if err != nil {
			// Handle specific error cases based on the returned error
			switch err {
			case utils.RepoErrUserNotFound:
				utils.ErrorResponse(c, utils.ErrUserNotFound)
				return
			case utils.RepoErrToWalletNotFound:
				utils.ErrorResponse(c, utils.ErrWalletNotFound)
				return
			case utils.RepoErrInsufficientFunds:
				utils.ErrorResponse(c, utils.ErrorInsufficientFunds)
				return
			case utils.ErrDatabaseError:
				utils.ErrorResponse(c, utils.ErrDatabaseError)
			default:
				utils.ErrorResponse(c, utils.ErrInternalServerError)
				return
			}
		}

		// Success response with the updated wallet balance
		utils.SuccessResponse(c, utils.MsgTransferSuccessful, gin.H{
			"balance":    wallet.Balance,
			"updated_at": wallet.UpdatedAt, // last updated time
		})
	}
}

// TransactionHistoryHandler returns the transaction history for the authenticated user
func TransactionHistoryHandler(ts transaction.TransactionServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user ID from the context (set by JWT middleware)
		userIDInterface, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, utils.ErrUnauthorized)
			return
		}

		// Ensure userID is of type int
		if _, ok := userIDInterface.(int); !ok {
			utils.ErrorResponse(c, utils.ErrInvalidUserId)
			return
		}

		// Get the wallet number from the context (set by WalletNumberMiddleware)
		walletNumberInterface, exists := c.Get("wallet_number")
		if !exists {
			utils.ErrorResponse(c, utils.ErrWalletNotFound)
			return
		}

		// Ensure walletNumber is of type string
		walletNumber, ok := walletNumberInterface.(string)
		if !ok {
			utils.ErrorResponse(c, utils.ErrorWalletNumber)
			return
		}

		// Parse query parameters for sorting and limiting
		orderBy := c.DefaultQuery("order", "desc")
		if orderBy != "asc" && orderBy != "desc" {
			utils.ErrorResponse(c, utils.ErrorInvalidOrder)
			return
		}

		// Limit query parameter
		const maxLimit = 100
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > maxLimit {
			utils.ErrorResponse(c, utils.ErrorInvalidLimit)
			return
		}

		// Offset query parameter
		offsetStr := c.DefaultQuery("offset", "0")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			utils.ErrorResponse(c, utils.ErrorInvalidOffset)
			return
		}

		// Get the transaction history using the wallet number
		transactions, err := ts.GetTransactionHistory(walletNumber, orderBy, limit, offset)
		if err != nil {
			utils.ErrorResponse(c, utils.ErrInternalServerError)
			return
		}

		// Return the transactions as JSON
		utils.SuccessResponse(c, utils.MsgTransactionRetrieved, gin.H{"transactions": transactions})
	}
}

func CreateWalletHandler(ws WalletServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from the context (set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, utils.ErrUnauthorized)
			return
		}

		// Create the wallet using the WalletService
		wallet, err := ws.CreateWallet(userID.(int))
		if err != nil {
			switch err {
			case utils.ErrDatabaseError:
				utils.ErrorResponse(c, utils.ErrDatabaseError)
				return
			case utils.ErrWalletAlreadyExists:
				utils.ErrorResponse(c, utils.ErrWalletAlreadyExists)
				return
			default:
				utils.ErrorResponse(c, utils.ErrInternalServerError)
				return
			}
		}

		// Return structured success response with wallet number
		utils.SuccessResponse(c, utils.MsgWalletCreated, gin.H{
			"wallet_number": wallet.WalletNumber,
		})
	}
}
