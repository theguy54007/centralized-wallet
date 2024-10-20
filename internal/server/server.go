package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"centralized-wallet/internal/auth"
	"centralized-wallet/internal/database"
	"centralized-wallet/internal/redis"
	"centralized-wallet/internal/transaction"
	"centralized-wallet/internal/user"
	"centralized-wallet/internal/wallet"
)

type Server struct {
	port int

	db                 database.Service
	rd                 redis.RedisService
	blackListService   *auth.BlacklistService
	userService        *user.UserService
	transactionService *transaction.TransactionService
	walletService      *wallet.WalletService
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	rd := redis.NewRedisService()
	dbService := database.New()
	// Initialize repositories
	userRepo := user.NewUserRepository(dbService.GetDB())
	walletRepo := wallet.NewWalletRepository(dbService.GetDB())
	transactionRepo := transaction.NewTransactionRepository(dbService.GetDB())

	// Initialize services

	transactionService := transaction.NewTransactionService(transactionRepo)
	walletService := wallet.NewWalletService(walletRepo, transactionService)
	userService := user.NewUserService(userRepo)
	NewServer := &Server{
		port: port,

		db:                 dbService,
		rd:                 *rd,
		blackListService:   auth.NewBlacklistService(rd),
		userService:        userService,
		walletService:      walletService,
		transactionService: transactionService,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
