package wallet

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/utils"

	"centralized-wallet/tests/testutils"
	"fmt"

	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// Balance Handler Test
func TestBalanceHandler(t *testing.T) {

	testRequest := testutils.TestHandlerRequest{
		Method: "GET",
		URL:    "/wallets/balance",
	}

	// Define the test cases
	testCases := []testWalletHandler{
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Successful balance retrieval",
				TestType: "success",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				MockSetup: func() {
					// Mock successful balance retrieval
					mockWallet := createMockWallet(testWalletNumber, testUserID)
					mockHandlerTestHelper.walletService.On("GetWalletByUserID", testUserID).
						Return(mockWallet, nil)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
				ExpectedStatus: http.StatusOK,
				ExpectedEntity: gin.H{
					"wallet_number": testWalletNumber,
					"balance":       100.0,
					"updated_at":    now.Format(time.RFC3339Nano),
				},
				ExpectedResponseError: nil,
				ExpectedMessage:       utils.MsgBalanceRetrieved,
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Wallet not found",
				TestType: "error",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				MockSetup: func() {
					// Mock wallet not found error
					mockHandlerTestHelper.walletService.On("GetWalletByUserID", testUserID).
						Return(nil, utils.RepoErrWalletNotFound)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
				ExpectedStatus:        http.StatusNotFound,
				ExpectedResponseError: utils.ErrWalletNotFound,
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Internal server error",
				TestType: "error",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				MockSetup: func() {
					// Mock an internal server error
					mockHandlerTestHelper.walletService.On("GetWalletByUserID", testUserID).
						Return(nil, utils.ErrInternalServerError)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
				ExpectedStatus:        http.StatusInternalServerError,
				ExpectedResponseError: utils.ErrInternalServerError,
			},
			userID: testUserID,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Call the common test flow handler
			walletHandlerTestFlow(tc, t)
		})
	}
}

// Deposit Handler Test
func TestDepositHandler(t *testing.T) {

	testRequest := testutils.TestHandlerRequest{
		Method: "POST",
		URL:    "/wallets/deposit",
	}

	// Define the test cases
	testCases := []testWalletHandler{
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Successful deposit",
				TestType: "success",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				Body: map[string]interface{}{
					"amount": testAmount,
				},
				MockSetup: func() {
					// Mock successful deposit
					mockHandlerTestHelper.walletService.On("Deposit", testUserID, testAmount).
						Return(&models.Wallet{
							UserID:    testUserID,
							Balance:   testAmount + 100.0, // Assume balance is updated after deposit
							UpdatedAt: now,
						}, nil)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
				ExpectedStatus: http.StatusOK,
				ExpectedEntity: gin.H{
					"balance":    testAmount + 100.0,
					"updated_at": now.Format(time.RFC3339Nano),
				},
				ExpectedResponseError: nil,
				ExpectedMessage:       utils.MsgDepositSuccessful,
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Wallet not found",
				TestType: "error",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				Body: map[string]interface{}{
					"amount": testAmount,
				},
				MockSetup: func() {
					// Mock wallet not found error
					mockHandlerTestHelper.walletService.On("Deposit", testUserID, testAmount).
						Return(nil, utils.RepoErrWalletNotFound)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
				ExpectedStatus:        http.StatusNotFound,
				ExpectedResponseError: utils.ErrWalletNotFound,
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Invalid request body",
				TestType: "error",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				Body: map[string]interface{}{
					"invalid_field": 12345, // Invalid field to simulate a malformed request
				},
				MockSetup: func() {
					// No service mock required for invalid request
				},
				MockAssert:            func(t *testing.T) {},
				ExpectedStatus:        http.StatusBadRequest,
				ExpectedResponseError: utils.ErrInvalidRequest,
			},
			userID: testUserID,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			// Call the common test flow handler
			walletHandlerTestFlow(tc, t)
		})
	}
}

// Withdraw Handler Test
func TestWithdrawHandler(t *testing.T) {

	testRequest := testutils.TestHandlerRequest{
		Method: "POST",
		URL:    "/wallets/withdraw",
	}

	// Define the test cases
	testCases := []testWalletHandler{
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Successful withdrawal",
				TestType: "success",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				Body: map[string]interface{}{
					"amount": testAmount,
				},
				MockSetup: func() {
					// Mock successful withdrawal
					mockHandlerTestHelper.walletService.On("Withdraw", testUserID, testAmount).
						Return(&models.Wallet{
							UserID:    testUserID,
							Balance:   50.0,
							UpdatedAt: now,
						}, nil)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
				ExpectedStatus: http.StatusOK,
				ExpectedEntity: gin.H{
					"balance":    50.0,
					"updated_at": now.Format(time.RFC3339Nano),
				},
				ExpectedResponseError: nil,
				ExpectedMessage:       utils.MsgWithdrawSuccessful,
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "User not found",
				TestType: "error",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				Body: map[string]interface{}{
					"amount": testAmount,
				},
				MockSetup: func() {
					// Mock user not found error
					mockHandlerTestHelper.walletService.On("Withdraw", testUserID, testAmount).
						Return(nil, utils.RepoErrUserNotFound)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
				ExpectedStatus:        http.StatusNotFound,
				ExpectedResponseError: utils.ErrUserNotFound,
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Insufficient funds",
				TestType: "error",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				Body: map[string]interface{}{
					"amount": testAmount,
				},
				MockSetup: func() {
					// Mock insufficient funds error
					mockHandlerTestHelper.walletService.On("Withdraw", testUserID, testAmount).
						Return(nil, utils.RepoErrInsufficientFunds)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
				ExpectedStatus:        http.StatusBadRequest,
				ExpectedResponseError: utils.ErrorInsufficientFunds,
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Invalid request body",
				TestType: "error",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				Body: map[string]interface{}{
					"invalid_field": -1.24, // Invalid field to simulate a malformed request
				},
				MockSetup: func() {
					// No need to mock any service since the error occurs before reaching the service layer
				},
				MockAssert:            func(t *testing.T) {},
				ExpectedStatus:        http.StatusBadRequest,
				ExpectedResponseError: utils.ErrInvalidRequest,
			},
			userID: testUserID,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Call the common test flow handler
			walletHandlerTestFlow(tc, t)
		})
	}
}

// Transaction History Test
func TestTransferHandler(t *testing.T) {

	testRequest := testutils.TestHandlerRequest{
		Method: "POST",
		URL:    "/wallets/transfer",
	}

	// Define the test cases
	testCases := []testWalletHandler{
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "ToUserNotExist",
				TestType: "error",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				Body: map[string]interface{}{
					"to_wallet_number": testToWalletNumber,
					"amount":           50.0,
				},
				ExpectedResponseError: utils.ErrWalletNotFound,
				MockSetup: func() {
					// Mock Transfer with user existence failure
					mockHandlerTestHelper.walletService.On("Transfer", testUserID, testToWalletNumber, 50.0).
						Return((*models.Wallet)(nil), utils.RepoErrWalletNotFound)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "FromUserNotExist",
				TestType: "error",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				Body: map[string]interface{}{
					"to_wallet_number": testToWalletNumber,
					"amount":           50.0,
				},
				ExpectedResponseError: utils.ErrUserNotFound,
				MockSetup: func() {
					// Mock Transfer with from_user_id failure
					mockHandlerTestHelper.walletService.On("Transfer", testUserID, testToWalletNumber, 50.0).
						Return((*models.Wallet)(nil), utils.RepoErrUserNotFound)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Success",
				TestType: "success",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				Body: map[string]interface{}{
					"to_wallet_number": testToWalletNumber,
					"amount":           50.0,
				},
				MockSetup: func() {
					// Mock successful transfer
					mockHandlerTestHelper.walletService.On("Transfer", testUserID, testToWalletNumber, 50.0).
						Return(&models.Wallet{
							UserID:    testUserID,
							Balance:   100.0,
							UpdatedAt: now,
						}, nil)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
					mockHandlerTestHelper.transactionSerivce.AssertExpectations(t)
				},
				ExpectedStatus:        http.StatusOK,
				ExpectedResponseError: nil,
				ExpectedMessage:       utils.MsgTransferSuccessful,
				ExpectedEntity: gin.H{
					"balance":    100.0,
					"updated_at": now.Format(time.RFC3339Nano),
				},
			},
			userID: testUserID,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			walletHandlerTestFlow(tc, t)
		})
	}
}

func TestCreateWalletHandler(t *testing.T) {

	testRequest := testutils.TestHandlerRequest{
		Method: "POST",
		URL:    "/wallets/create",
	}

	// Define the test cases
	testCases := []testWalletHandler{
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "successful wallet creation",
				TestType: "success",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				MockSetup: func() {
					// Mock wallet creation success
					mockHandlerTestHelper.walletService.On("CreateWallet", testUserID).
						Return(&models.Wallet{
							WalletNumber: testWalletNumber,
							UserID:       testUserID,
							Balance:      0.0,
							UpdatedAt:    time.Now(),
						}, nil)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
				ExpectedStatus:  http.StatusOK,
				ExpectedMessage: utils.MsgWalletCreated,
				ExpectedEntity: gin.H{
					"wallet_number": testWalletNumber,
				},
				ExpectedResponseError: nil,
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "user already has a wallet",
				TestType: "error",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				MockSetup: func() {
					// Mock wallet already exists error
					mockHandlerTestHelper.walletService.On("CreateWallet", testUserID).
						Return(nil, utils.ErrWalletAlreadyExists)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
				ExpectedStatus:        http.StatusConflict,
				ExpectedResponseError: utils.ErrWalletAlreadyExists,
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "unknown internal error",
				TestType: "error",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				MockSetup: func() {
					// Mock unknown error
					mockHandlerTestHelper.walletService.On("CreateWallet", testUserID).
						Return(nil, fmt.Errorf("some random error"))
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.walletService.AssertExpectations(t)
				},
				ExpectedStatus:        http.StatusInternalServerError,
				ExpectedResponseError: utils.ErrInternalServerError,
			},
			userID: testUserID,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Call the common test flow handler
			walletHandlerTestFlow(tc, t)
		})
	}
}

func TestTransactionHistoryHandler(t *testing.T) {

	testRequest := testutils.TestHandlerRequest{
		Method: "GET",
		URL:    "/wallets/transactions?order=desc&limit=10",
	}

	formatTransactions := []models.FormattedTransaction{
		{
			TransactionType: "deposit",
			Amount:          100.0,
			Direction:       "incoming",
		},
		{
			TransactionType: "withdraw",
			Amount:          testAmount,
			Direction:       "outgoing",
		},
		{
			TransactionType: "transfer",
			Amount:          testAmount,
			Direction:       "outgoing",
			ToWalletNumber:  testToWalletNumber,
			ToEmail:         toTestEmail,
		},
		{
			TransactionType:  "transfer",
			Amount:           testAmount,
			Direction:        "incoming",
			FromWalletNumber: testToWalletNumber,
			FromEmail:        toTestEmail,
		},
	}

	// Define the test cases
	testCases := []testWalletHandler{
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Successful transaction history retrieval",
				TestType: "success",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				MockSetup: func() {
					// Mock successful transaction history retrieval
					mockHandlerTestHelper.transactionSerivce.On("GetTransactionHistory", testFromWalletNumber, "desc", 10, 0).
						Return(formatTransactions, nil)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.transactionSerivce.AssertExpectations(t)
				},
				ExpectedStatus: http.StatusOK,
				ExpectedEntity: gin.H{
					"wallet_number": testFromWalletNumber,
					"transactions":  formatTransactions,
				},
				ExpectedResponseError: nil,
				ExpectedMessage:       utils.MsgTransactionRetrieved,
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "No transactions for wallet",
				TestType: "success",
				URL:      testRequest.URL,
				Method:   testRequest.Method,
				MockSetup: func() {
					// Mock no transactions found (empty array)
					mockHandlerTestHelper.transactionSerivce.On("GetTransactionHistory", testFromWalletNumber, "desc", 10, 0).
						Return([]models.FormattedTransaction{}, nil)
				},
				MockAssert: func(t *testing.T) {
					mockHandlerTestHelper.transactionSerivce.AssertExpectations(t)
				},
				ExpectedStatus: http.StatusOK,
				ExpectedEntity: gin.H{
					"wallet_number": testFromWalletNumber,
					"transactions":  []models.FormattedTransaction{}, // Expecting an empty array
				},
				ExpectedResponseError: nil,
				ExpectedMessage:       utils.MsgTransactionRetrieved,
			},
			userID: testUserID,
		},
		{
			BaseHandlerTestCase: testutils.BaseHandlerTestCase{
				Name:     "Invalid query parameters (limit)",
				TestType: "error",
				URL:      "/wallets/transactions?order=desc&limit=0", // Invalid limit
				Method:   testRequest.Method,
				MockSetup: func() {
					// No need to mock service for invalid request
				},
				MockAssert:            func(t *testing.T) {},
				ExpectedStatus:        http.StatusBadRequest,
				ExpectedResponseError: utils.ErrorInvalidLimit,
			},
			userID: testUserID,
		},
	}

	// Iterate over the test cases
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {

			// Call the common test flow handler
			walletHandlerTestFlow(tc, t)
		})
	}
}
