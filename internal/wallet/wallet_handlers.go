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
			utils.ErrorResponse(c, utils.ErrUnauthorized, nil, "")
			return
		}

		// Fetch balance from the WalletService
		wallet, err := ws.GetWalletByUserID(userID.(int))
		if err != nil {
			// Handle specific error cases
			switch err {
			case utils.RepoErrWalletNotFound:
				utils.ErrorResponse(c, utils.ErrWalletNotFound, nil, "")
			default:
				utils.ErrorResponse(c, utils.ErrInternalServerError, err, "[BalanceHandler] Error getting wallet by user ID")
			}
			return
		}

		// Respond with balance
		utils.SuccessResponse(c, utils.MsgBalanceRetrieved, gin.H{
			"wallet_number": wallet.WalletNumber,
			"balance":       wallet.Balance,
			"updated_at":    wallet.UpdatedAt, // timestamp of last wallet update
		})
	}
}

// DepositHandler handles deposit requests and returns the updated balance and updated_at time
func DepositHandler(ws WalletServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from the context (set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, utils.ErrUnauthorized, nil, "")
			return
		}

		// Parse request body
		var request struct {
			Amount float64 `json:"amount" binding:"required,gt=0"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			utils.ErrorResponse(c, utils.ErrInvalidRequest, nil, "")
			return
		}

		// Perform the deposit and get the updated Wallet struct
		wallet, err := ws.Deposit(userID.(int), request.Amount)
		if err != nil {
			switch err {
			case utils.RepoErrWalletNotFound:
				utils.ErrorResponse(c, utils.ErrWalletNotFound, nil, "")
			default:
				utils.ErrorResponse(c, utils.ErrInternalServerError, err, "[DepositHandler] Error depositing amount")
			}
			return
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
			utils.ErrorResponse(c, utils.ErrUnauthorized, nil, "")
			return
		}

		// Parse the request body
		var request struct {
			Amount float64 `json:"amount" binding:"required,gt=0"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			utils.ErrorResponse(c, utils.ErrInvalidRequest, nil, "")
			return
		}

		// Perform the withdrawal and get the updated Wallet struct
		wallet, err := ws.Withdraw(userID.(int), request.Amount)
		if err != nil {
			switch err {
			case utils.RepoErrUserNotFound:
				utils.ErrorResponse(c, utils.ErrUserNotFound, nil, "")
			case utils.RepoErrInsufficientFunds:
				utils.ErrorResponse(c, utils.ErrorInsufficientFunds, nil, "")
			default:
				utils.ErrorResponse(c, utils.ErrInternalServerError, err, "[WithdrawHandler] Error withdrawing amount")
			}
			return
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
			utils.ErrorResponse(c, utils.ErrUnauthorized, nil, "")
			return
		}

		// Parse and validate request payload
		var request struct {
			ToWalletNumber string  `json:"to_wallet_number" binding:"required"`
			Amount         float64 `json:"amount" binding:"required,gt=0"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			utils.ErrorResponse(c, utils.ErrInvalidRequest, nil, "")
			return
		}

		// Perform the transfer operation
		wallet, err := ws.Transfer(fromUserID.(int), request.ToWalletNumber, request.Amount)
		if err != nil {
			// Handle specific error cases based on the returned error
			switch err {
			case utils.RepoErrUserNotFound:
				utils.ErrorResponse(c, utils.ErrUserNotFound, nil, "")
			case utils.RepoErrWalletNotFound:
				utils.ErrorResponse(c, utils.ErrWalletNotFound, nil, "")
			case utils.RepoErrInsufficientFunds:
				utils.ErrorResponse(c, utils.ErrorInsufficientFunds, nil, "")
			default:
				utils.ErrorResponse(c, utils.ErrInternalServerError, err, "[TransferHandler] Error transferring amount")
			}
			return
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
			utils.ErrorResponse(c, utils.ErrUnauthorized, nil, "")
			return
		}

		// Ensure userID is of type int
		if _, ok := userIDInterface.(int); !ok {
			utils.ErrorResponse(c, utils.ErrInvalidUserId, nil, "")
			return
		}

		// Get the wallet number from the context (set by WalletNumberMiddleware)
		walletNumberInterface, exists := c.Get("wallet_number")
		if !exists {
			utils.ErrorResponse(c, utils.ErrWalletNotFound, nil, "")
			return
		}

		// Ensure walletNumber is of type string
		walletNumber, ok := walletNumberInterface.(string)
		if !ok {
			utils.ErrorResponse(c, utils.ErrorWalletNumber, nil, "")
			return
		}

		// Parse query parameters for sorting and limiting
		orderBy := c.DefaultQuery("order", "desc")
		if orderBy != "asc" && orderBy != "desc" {
			utils.ErrorResponse(c, utils.ErrorInvalidOrder, nil, "")
			return
		}

		// Limit query parameter
		const maxLimit = 100
		limitStr := c.DefaultQuery("limit", "10")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > maxLimit {
			utils.ErrorResponse(c, utils.ErrorInvalidLimit, nil, "")
			return
		}

		// Offset query parameter
		offsetStr := c.DefaultQuery("offset", "0")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			utils.ErrorResponse(c, utils.ErrorInvalidOffset, nil, "")
			return
		}

		// Get the transaction history using the wallet number
		transactions, err := ts.GetTransactionHistory(walletNumber, orderBy, limit, offset)
		if err != nil {
			utils.ErrorResponse(c, utils.ErrInternalServerError, err, "[TransactionHistoryHandler] Error getting transaction history")
			return
		}

		// Return the transactions as JSON
		utils.SuccessResponse(c, utils.MsgTransactionRetrieved, gin.H{"wallet_number": walletNumber, "transactions": transactions})
	}
}

func CreateWalletHandler(ws WalletServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from the context (set by JWTMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			utils.ErrorResponse(c, utils.ErrUnauthorized, nil, "")
			return
		}

		// Create the wallet using the WalletService
		wallet, err := ws.CreateWallet(userID.(int))
		if err != nil {
			switch err {
			case utils.ErrWalletAlreadyExists:
				utils.ErrorResponse(c, utils.ErrWalletAlreadyExists, nil, "")
			default:
				utils.ErrorResponse(c, utils.ErrInternalServerError, err, "[CreateWalletHandler] Error creating wallet")
			}
			return
		}

		// Return structured success response with wallet number
		utils.SuccessResponse(c, utils.MsgWalletCreated, gin.H{
			"wallet_number": wallet.WalletNumber,
		})
	}
}
