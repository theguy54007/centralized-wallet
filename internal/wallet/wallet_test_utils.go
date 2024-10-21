package wallet

import (
	"centralized-wallet/internal/auth"
	"centralized-wallet/internal/models"
	mockAuth "centralized-wallet/tests/mocks/auth"
	mockRedis "centralized-wallet/tests/mocks/redis"
	mockTransaction "centralized-wallet/tests/mocks/transaction"
	mockWallet "centralized-wallet/tests/mocks/wallet"
	"centralized-wallet/tests/testutils"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// Shared test variables and mock data
var (
	testUserID           = 1
	testToUserID         = 2
	testToWalletNumber   = "1234567890"
	testFromWalletNumber = "0987654321"
	testAmount           = 50.0
	testEmail            = "user1@example.com"
	toTestEmail          = "user1@example.com"
	now                  = time.Now()
	testWalletNumber     = "1234567890"
)

// wallet handler test case struct
type testWalletHandler struct {
	testutils.BaseHandlerTestCase
	userID int
}

var mockHandlerTestHelper struct {
	transactionSerivce *mockTransaction.MockTransactionService
	walletService      *mockWallet.MockWalletService
	blacklistService   *mockAuth.MockBlacklistService
	redisClient        *mockRedis.MockRedisClient
}

func setupHandlerMock() {
	mockHandlerTestHelper.transactionSerivce = new(mockTransaction.MockTransactionService)
	mockHandlerTestHelper.walletService = new(mockWallet.MockWalletService)
	mockHandlerTestHelper.blacklistService = new(mockAuth.MockBlacklistService)
	mockHandlerTestHelper.redisClient = new(mockRedis.MockRedisClient)

	mockHandlerTestHelper.redisClient.On("Get", mock.Anything, "user:1:wallet_number").Return(testFromWalletNumber, nil)
}

func generateJWTForTest(userID int) string {
	token, _ := auth.GenerateJWT(userID)
	return token
}

func walletHandlerTestFlow(tc testWalletHandler, t *testing.T) {
	router := setupHandlerRouter()

	tc.MockSetup()

	token := generateJWTForTest(tc.userID)

	w := testutils.ExecuteRequest(router, tc.Method, tc.URL, tc.Body, token)

	// Assert status code and response body
	if tc.TestType == "success" {
		testutils.AssertAPISuccessResponse(t, w, tc.ExpectedMessage, tc.ExpectedEntity, tc.ExpectedStatus)
	} else if tc.TestType == "error" {
		testutils.AssertAPIErrorResponse(t, w, tc.ExpectedResponseError)
	}
}

func setupHandlerRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	setupHandlerMock()

	// should run every time a request is made
	mockHandlerTestHelper.blacklistService.On("IsTokenBlacklisted", generateJWTForTest(testUserID)).Return(false, nil)

	walletRoutes := router.Group("/wallets")
	walletRoutes.Use(auth.JWTMiddleware(mockHandlerTestHelper.blacklistService))
	{
		walletRoutes.GET("/balance", BalanceHandler(mockHandlerTestHelper.walletService))
		walletRoutes.POST("/deposit", DepositHandler(mockHandlerTestHelper.walletService))
		walletRoutes.POST("/withdraw", WithdrawHandler(mockHandlerTestHelper.walletService))
		walletRoutes.POST("/transfer", TransferHandler(mockHandlerTestHelper.walletService))
		walletRoutes.POST("/create", CreateWalletHandler(mockHandlerTestHelper.walletService))
		walletRoutes.GET("/transactions",
			WalletNumberMiddleware(mockHandlerTestHelper.walletService, mockHandlerTestHelper.redisClient),
			TransactionHistoryHandler(mockHandlerTestHelper.transactionSerivce),
		)
	}
	return router
}

func createMockWallet(walletNumber string, userId int) *models.Wallet {
	return &models.Wallet{
		UserID:       userId,
		Balance:      100.0,
		WalletNumber: walletNumber,
		UpdatedAt:    now,
	}
}

// wallet service test case struct
type testWalletService struct {
	testutils.BaseHandlerTestCase
	userID       int
	walletNumber string
	amount       float64
}

func walletServiceTestInit(tt testWalletService) WalletServiceInterface {
	setupServiceMock()
	tt.MockSetup()
	return NewWalletService(mockServiceTestHelper.walletRepo, mockServiceTestHelper.transactionService)
}

func setupServiceMock() {
	mockServiceTestHelper.walletRepo = new(mockWallet.MockWalletRepository)
	mockServiceTestHelper.transactionService = new(mockTransaction.MockTransactionService)
}

var mockServiceTestHelper struct {
	walletRepo         *mockWallet.MockWalletRepository
	transactionService *mockTransaction.MockTransactionService
}
