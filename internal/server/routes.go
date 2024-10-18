package server

import (
	"net/http"

	"centralized-wallet/internal/repository"
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

	// Initialize the WalletRepository and WalletService
	// Initialize WalletService
	walletRepo := repository.NewWalletRepository(s.db.GetDB()) // Assuming walletRepo is initialized in Server
	walletService := wallet.NewWalletService(walletRepo)

	// Wallet routes
	r.GET("/wallets/:user_id/balance", wallet.BalanceHandler(walletService))
	r.POST("/wallets/:user_id/deposit", wallet.DepositHandler(walletService))
	r.POST("/wallets/:user_id/withdraw", wallet.WithdrawHandler(walletService))
	r.POST("/wallets/transfer", wallet.TransferHandler(walletService))

	return r
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}
