package wallet

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/transaction"
	"centralized-wallet/internal/utils"
	"database/sql"
	"fmt"
	"log"

	"math/rand"
	"time"
)

// WalletServiceInterface defines the methods for the WalletService
type WalletServiceInterface interface {
	// GetBalance(userID int) (float64, error)
	UserExists(userID int) (bool, error)
	Deposit(userID int, amount float64) (*models.Wallet, error)
	Withdraw(userID int, amount float64) (*models.Wallet, error)
	Transfer(fromUserID int, toWalletNumber string, amount float64) (*models.Wallet, error)
	GetWalletByUserID(userID int) (*models.Wallet, error)
	CreateWallet(userID int) (*models.Wallet, error)
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

func (ws *WalletService) CreateWallet(userID int) (*models.Wallet, error) {
	// check the user is already exists
	userExists, err := ws.UserExists(userID)
	if err != nil {
		return nil, err
	}
	// if user already exists, return error
	if userExists {
		return nil, utils.ErrWalletAlreadyExists
	}

	walletNumber := generateUniqueWalletNumber(userID)
	wallet := &models.Wallet{
		UserID:       userID,
		Balance:      0.0,
		WalletNumber: walletNumber,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Call repository to insert wallet in the database using the transaction
	err = ws.walletRepo.CreateWallet(wallet)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

func (ws *WalletService) UserExists(userID int) (bool, error) {
	return ws.walletRepo.UserExists(userID)
}

// Deposit adds money to the user's wallet and records the transaction, returning balance and timestamp
func (ws *WalletService) Deposit(userID int, amount float64) (*models.Wallet, error) {
	exists, err := ws.walletRepo.UserExists(userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		// Return a specific error if the wallet is not found
		return nil, utils.RepoErrWalletNotFound
	}

	tx, err := ws.walletRepo.Begin()
	if err != nil {
		return nil, err
	}

	// Rollback the transaction if an error occurs
	defer ws.rollBackTxWhenErr(tx, &err)

	// Perform the deposit and get the updated Wallet struct
	wallet, err := ws.walletRepo.Deposit(tx, userID, amount)
	if err != nil {
		return nil, err
	}

	// Record the deposit transaction
	err = ws.transactionService.RecordTransaction(tx, nil, &wallet.WalletNumber, "deposit", amount)
	if err != nil {
		return nil, err
	}

	err = ws.walletRepo.Commit(tx)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// Withdraw subtracts money from the user's wallet, records the transaction, and returns updated balance and updated_at time
func (ws *WalletService) Withdraw(userID int, amount float64) (*models.Wallet, error) {
	checkWallet, err := ws.walletRepo.GetWalletByUserID(userID)
	if err != nil {
		return nil, err
	}

	if checkWallet.Balance < amount {
		return nil, utils.RepoErrInsufficientFunds
	}

	tx, err := ws.walletRepo.Begin()
	if err != nil {
		return nil, err
	}

	defer ws.rollBackTxWhenErr(tx, &err)

	// Withdraw and get the updated wallet data
	wallet, err := ws.walletRepo.Withdraw(tx, userID, amount)
	if err != nil {
		return nil, err
	}

	// Record the withdrawal transaction
	err = ws.transactionService.RecordTransaction(tx, &wallet.WalletNumber, nil, "withdraw", amount)
	if err != nil {
		return nil, err
	}

	err = ws.walletRepo.Commit(tx)
	if err != nil {
		return nil, err
	}

	// Return the updated wallet (including balance and updated_at)
	return wallet, nil
}

// Transfer subtracts from one user and adds to another, returning the updated Wallet for the from_user
func (ws *WalletService) Transfer(fromUserID int, toWalletNumber string, amount float64) (*models.Wallet, error) {

	checkWallet, err := ws.walletRepo.GetWalletByUserID(fromUserID)
	if err != nil {
		log.Printf("Error getting wallet balance: %v", err)
		return nil, err
	}

	if checkWallet.Balance < amount {
		return nil, utils.RepoErrInsufficientFunds
	}

	toWallet, err := ws.walletRepo.FindByWalletNumber(toWalletNumber)
	if err != nil {
		return nil, err
	}

	tx, err := ws.walletRepo.Begin()
	if err != nil {
		return nil, err
	}

	defer ws.rollBackTxWhenErr(tx, &err)

	fromWallet, err := ws.walletRepo.Withdraw(tx, fromUserID, amount)
	if err != nil {
		return nil, err
	}

	toWallet, err = ws.walletRepo.Deposit(tx, toWallet.UserID, amount)
	if err != nil {
		return nil, err
	}

	// Record the transfer transaction
	err = ws.transactionService.RecordTransaction(tx, &fromWallet.WalletNumber, &toWallet.WalletNumber, "transfer", amount)
	if err != nil {
		return nil, err
	}

	err = ws.walletRepo.Commit(tx)
	if err != nil {
		return nil, err
	}

	return fromWallet, nil
}

func (ws *WalletService) rollBackTxWhenErr(tx *sql.Tx, err *error) {
	if err != nil {
		ws.walletRepo.Rollback(tx)
	}
}

func generateUniqueWalletNumber(userID int) string {
	// Get the current timestamp in the format YYYYMMDDHHMMSS
	timestamp := time.Now().Format("20060102150405")

	if len(timestamp) > 3 {
		timestamp = timestamp[3:] // Remove the first three digits
	}
	// Generate a small random string
	randomString := generateRandomString(6) // Example: ABC123

	// Combine userID, timestamp, and random string into a wallet number
	walletNumber := fmt.Sprintf("WAL-%d-%s-%s", userID, timestamp, randomString)

	return walletNumber
}

// generateRandomString generates a random string of length n.
func generateRandomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano())) // Create a new random source
	result := make([]byte, n)
	for i := range result {
		result[i] = letters[seededRand.Intn(len(letters))] // Use the new seeded random instance
	}
	return string(result)
}
