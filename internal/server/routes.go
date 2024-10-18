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

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.GET("/health", s.healthHandler)

	// Initialize UserRepository and UserService
	userRepo := repository.NewUserRepository(s.db.GetDB())
	userService := user.NewUserService(userRepo)

	// User routes
	r.POST("/register", user.RegistrationHandler(userService))
	r.POST("/login", user.LoginHandler(userService))

	walletRepo := repository.NewWalletRepository(s.db.GetDB()) // Ensure this implements WalletRepositoryInterface
	transactionRepo := repository.NewTransactionRepository(s.db.GetDB())

	transactionService := transaction.NewTransactionService(transactionRepo) // Initialize TransactionService
	walletService := wallet.NewWalletService(walletRepo, transactionService) // Pass WalletRepositoryInterface

	// Wallet routes protected by JWT
	walletRoutes := r.Group("/wallets")
	walletRoutes.Use(middleware.JWTMiddleware()) // Apply JWT middleware to wallet routes
	{
		walletRoutes.GET("/balance", wallet.BalanceHandler(walletService))                      // Get balance
		walletRoutes.POST("/deposit", wallet.DepositHandler(walletService))                     // Deposit money
		walletRoutes.POST("/withdraw", wallet.WithdrawHandler(walletService))                   // Withdraw money
		walletRoutes.POST("/transfer", wallet.TransferHandler(walletService))                   // Transfer money
		walletRoutes.GET("/transactions", wallet.TransactionHistoryHandler(transactionService)) // transaction history
	}

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
