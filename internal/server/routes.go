package server

import (
	"net/http"

	"centralized-wallet/internal/auth"
	"centralized-wallet/internal/transaction"
	"centralized-wallet/internal/user"
	"centralized-wallet/internal/utils"
	"centralized-wallet/internal/wallet"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes initializes all routes for the application
func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(utils.ErrorMiddlewareHandler()) // Apply error middleware to all routes

	// Health check route
	r.GET("/db-health", s.dbHealthHandler)
	r.GET("/redis-health", s.redisHealthHandler)

	// Register all routes
	s.registerUserRoutes(r, s.userService)
	s.registerWalletRoutes(r, s.walletService, s.transactionService)

	return r
}

// healthHandler returns the health status of the application
func (s *Server) dbHealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

// healthHandler returns the health status of the application
func (s *Server) redisHealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.rd.Health(c))
}

// registerUserRoutes registers all routes related to users
func (s *Server) registerUserRoutes(r *gin.Engine, userService *user.UserService) {
	r.POST("/register", user.RegistrationHandler(userService))
	r.POST("/login", user.LoginHandler(userService))

	r.Use(auth.JWTMiddleware(s.blackListService)) // Apply JWT middleware to all user routes
	r.POST("/logout", user.LogoutHandler(s.blackListService))
}

// registerWalletRoutes registers all routes related to wallets and transactions
func (s *Server) registerWalletRoutes(r *gin.Engine, walletService *wallet.WalletService, transactionService transaction.TransactionServiceInterface) {
	walletRoutes := r.Group("/wallets")
	walletRoutes.Use(auth.JWTMiddleware(s.blackListService)) // Apply JWT middleware to all wallet routes

	walletRoutes.GET("/balance", wallet.BalanceHandler(walletService))    // Get balance
	walletRoutes.POST("/deposit", wallet.DepositHandler(walletService))   // Deposit money
	walletRoutes.POST("/withdraw", wallet.WithdrawHandler(walletService)) // Withdraw money
	walletRoutes.POST("/transfer", wallet.TransferHandler(walletService))
	walletRoutes.POST("/create", wallet.CreateWalletHandler(walletService))

	walletRoutes.Use(wallet.WalletNumberMiddleware(s.walletService, &s.rd))

	walletRoutes.GET("/transactions", wallet.TransactionHistoryHandler(transactionService)) // transaction history
}
