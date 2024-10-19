package server

import (
	"net/http"

	"centralized-wallet/internal/middleware"
	"centralized-wallet/internal/repository"
	"centralized-wallet/internal/transaction"
	"centralized-wallet/internal/user"
	"centralized-wallet/internal/wallet"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes initializes all routes for the application
func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	// Health check route
	r.GET("/db-health", s.dbHealthHandler)
	r.GET("/redis-health", s.redisHealthHandler)

	// Initialize repositories
	userRepo := repository.NewUserRepository(s.db.GetDB())
	walletRepo := repository.NewWalletRepository(s.db.GetDB())
	transactionRepo := repository.NewTransactionRepository(s.db.GetDB())

	// Initialize services
	userService := user.NewUserService(userRepo)
	transactionService := transaction.NewTransactionService(transactionRepo)
	walletService := wallet.NewWalletService(walletRepo, transactionService)

	// Register all routes
	s.registerUserRoutes(r, userService)
	s.registerWalletRoutes(r, walletService, transactionService)

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
}

// registerWalletRoutes registers all routes related to wallets and transactions
func (s *Server) registerWalletRoutes(r *gin.Engine, walletService *wallet.WalletService, transactionService transaction.TransactionServiceInterface) {
	walletRoutes := r.Group("/wallets")
	walletRoutes.Use(middleware.JWTMiddleware()) // Apply JWT middleware to all wallet routes

	walletRoutes.GET("/balance", wallet.BalanceHandler(walletService))                      // Get balance
	walletRoutes.POST("/deposit", wallet.DepositHandler(walletService))                     // Deposit money
	walletRoutes.POST("/withdraw", wallet.WithdrawHandler(walletService))                   // Withdraw money
	walletRoutes.POST("/transfer", wallet.TransferHandler(walletService))                   // Transfer money
	walletRoutes.GET("/transactions", wallet.TransactionHistoryHandler(transactionService)) // transaction history
}
