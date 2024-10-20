package server

import (
	"bytes"
	"centralized-wallet/internal/auth"
	"centralized-wallet/internal/database"
	"centralized-wallet/internal/transaction"
	"centralized-wallet/internal/user"
	"centralized-wallet/internal/wallet"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"centralized-wallet/tests/testutils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var dbService database.Service
var teardown func(context.Context) error

// TestMain sets up the test database and tears it down after all tests have completed.
func TestMain(m *testing.M) {
	// Start the PostgreSQL container
	var err error
	teardown, err = testutils.StartPostgresContainer(true)
	if err != nil {
		log.Fatalf("Could not start postgres container: %v", err)
	}

	// log.Print("routes_test testutils", testutils)
	testutils.InitEnv()
	// Initialize database
	dbService = database.New()

	// Run tests
	code := m.Run()

	// Tear down the container
	if teardown != nil && teardown(context.Background()) != nil {
		log.Fatalf("Could not teardown postgres container: %v", err)
	}

	// Exit with the proper code
	os.Exit(code)
}

// Mock JWT generation for tests
func generateJWTForTest(userId int) string {
	token, _ := auth.GenerateJWT(userId)
	return token
}

// Setup the router with real repositories and services using the test container DB
func setupRouterWithRealDB() *gin.Engine {
	// Initialize repositories and services
	userRepo := user.NewUserRepository(dbService.GetDB())
	walletRepo := wallet.NewWalletRepository(dbService.GetDB())
	transactionRepo := transaction.NewTransactionRepository(dbService.GetDB())

	transactionService := transaction.NewTransactionService(transactionRepo)
	walletService := wallet.NewWalletService(walletRepo, transactionService)
	userService := user.NewUserService(userRepo)
	// Initialize router
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Create Server instance and register routes
	serverInstance := &Server{db: dbService}
	serverInstance.registerUserRoutes(r, userService)
	serverInstance.registerWalletRoutes(r, walletService, transactionService)

	return r
}

// Test health route
// func TestHealthRoute(t *testing.T) {
// 	r := setupRouterWithRealDB()
// 	req, _ := http.NewRequest("GET", "/health", nil)
// 	w := httptest.NewRecorder()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.JSONEq(t, `{"status": "up"}`, w.Body.String())
// }

// Test user registration route
func TestUserRegistrationRoute(t *testing.T) {
	r := setupRouterWithRealDB()

	// Simulate registration request
	userData := `{"email":"test@example.com","password":"testpassword"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer([]byte(userData)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

// Test login route
// func TestLoginRoute(t *testing.T) {
// 	r := setupRouterWithRealDB()

// 	// Simulate login request
// 	userData := `{"email":"test@example.com","password":"testpassword"}`
// 	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(userData)))
// 	req.Header.Set("Content-Type", "application/json")
// 	w := httptest.NewRecorder()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	assert.Contains(t, w.Body.String(), `"token":"`)
// }

// // Test wallet routes protected by JWT
// func TestWalletRoutesProtectedByJWT(t *testing.T) {
// 	r := setupRouterWithRealDB()

// 	// Simulate a JWT-protected request
// 	token := generateJWTForTest(1)
// 	req, _ := http.NewRequest("GET", "/wallets/balance", nil)
// 	req.Header.Set("Authorization", "Bearer "+token)
// 	w := httptest.NewRecorder()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	expectedResponse := `{"balance":0}` // Assuming initial balance is 0
// 	assert.JSONEq(t, expectedResponse, w.Body.String())
// }

// // Test transaction history route
// func TestTransactionHistory(t *testing.T) {
// 	r := setupRouterWithRealDB()

// 	// Simulate a JWT-protected request
// 	token := generateJWTForTest(1)
// 	req, _ := http.NewRequest("GET", "/wallets/transactions", nil)
// 	req.Header.Set("Authorization", "Bearer "+token)
// 	w := httptest.NewRecorder()
// 	r.ServeHTTP(w, req)

// 	assert.Equal(t, http.StatusOK, w.Code)
// 	expectedResponse := `[]` // Assuming no transactions initially
// 	assert.JSONEq(t, expectedResponse, w.Body.String())
// }

// // Mock JWT middleware
// func JWTMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		token := c.GetHeader("Authorization")
// 		if token != "Bearer mock-valid-jwt-token" {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
// 			return
// 		}
// 		c.Set("user_id", 1) // Simulate valid user ID 1
// 		c.Next()
// 	}
// }
