package wallet

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/transaction"
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

// WalletServiceInterface defines the methods for the WalletService
type WalletServiceInterface interface {
	GetBalance(userID int) (float64, error)
	UserExists(userID int) (bool, error)
	Deposit(userID int, amount float64) (*models.Wallet, error)
	Withdraw(userID int, amount float64) (*models.Wallet, error)
	Transfer(fromUserID int, toWalletNumber string, amount float64) (*models.Wallet, error)
	GetWalletByUserID(userID int) (*models.Wallet, error)
	CreateWalletWithTx(tx *sql.Tx, userID int) (*models.Wallet, error)
}

// WalletService handles wallet operations using the repository interface
type WalletService struct {
	walletRepo         WalletRepositoryInterface
	transactionService transaction.TransactionServiceInterface
}

// GetWalletByUserID fetches the wallet by the user ID
func (ws *WalletService) GetWalletByUserID(userID int) (*models.Wallet, error) {
	return ws.walletRepo.GetWalletByUserID(userID)
}

// NewWalletService creates a new WalletService with the provided repository
func NewWalletService(walletRepo WalletRepositoryInterface, transactionService transaction.TransactionServiceInterface) *WalletService {
	return &WalletService{walletRepo: walletRepo, transactionService: transactionService}
}

func (ws *WalletService) CreateWalletWithTx(tx *sql.Tx, userID int) (*models.Wallet, error) {

	walletNumber := ws.GenerateUniqueWalletNumber(userID)
	wallet := &models.Wallet{
		UserID:       userID,
		Balance:      0.0,
		WalletNumber: walletNumber,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Call repository to insert wallet in the database using the transaction
	err := ws.walletRepo.CreateWalletWithTx(tx, wallet)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// GetBalance retrieves the balance of a user
func (ws *WalletService) GetBalance(userID int) (float64, error) {
	return ws.walletRepo.GetWalletBalance(userID)
}

func (ws *WalletService) UserExists(userID int) (bool, error) {
	return ws.walletRepo.UserExists(userID)
}

// Deposit adds money to the user's wallet and records the transaction, returning balance and timestamp
func (ws *WalletService) Deposit(userID int, amount float64) (*models.Wallet, error) {
	// Perform the deposit and get the updated Wallet struct
	wallet, err := ws.walletRepo.Deposit(userID, amount)
	if err != nil {
		return nil, err
	}

	// Record the deposit transaction
	err = ws.transactionService.RecordTransaction(nil, &wallet.WalletNumber, "deposit", amount)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// Withdraw subtracts money from the user's wallet and records the transaction
// Withdraw subtracts money from the user's wallet, records the transaction, and returns updated balance and updated_at time
func (ws *WalletService) Withdraw(userID int, amount float64) (*models.Wallet, error) {
	// Withdraw and get the updated wallet data
	wallet, err := ws.walletRepo.Withdraw(userID, amount)
	if err != nil {
		return nil, err
	}

	// Record the withdrawal transaction
	err = ws.transactionService.RecordTransaction(&wallet.WalletNumber, nil, "withdraw", amount)
	if err != nil {
		return nil, err
	}

	// Return the updated wallet (including balance and updated_at)
	return wallet, nil
}

// Transfer subtracts from one user and adds to another, returning the updated Wallet for the from_user
func (ws *WalletService) Transfer(fromUserID int, toWalletNumber string, amount float64) (*models.Wallet, error) {
	// Perform the transfer using the wallet number
	wallet, err := ws.walletRepo.Transfer(fromUserID, toWalletNumber, amount)
	if err != nil {
		return wallet, err
	}

	// Record the transfer transaction using user IDs (still retrieving the user ID for both wallets)
	toWallet, err := ws.walletRepo.FindByWalletNumber(toWalletNumber)
	if err != nil {
		return wallet, err
	}

	// Record the transfer transaction
	err = ws.transactionService.RecordTransaction(&wallet.WalletNumber, &toWallet.WalletNumber, "transfer", amount)
	if err != nil {
		return wallet, err
	}

	return wallet, nil
}

func (ws *WalletService) GenerateUniqueWalletNumber(userID int) string {
	// Get the current timestamp in the format YYYYMMDDHHMMSS
	timestamp := time.Now().Format("20060102150405")

	// Generate a small random string
	randomString := generateRandomString(6) // Example: ABC123

	// Combine userID, timestamp, and random string into a wallet number
	walletNumber := fmt.Sprintf("WAL-%d-%s-%s", userID, timestamp, randomString)

	return walletNumber
}

// generateRandomString generates a random string of length n.
func generateRandomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}
