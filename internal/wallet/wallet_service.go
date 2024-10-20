package wallet

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/transaction"
)

// WalletServiceInterface defines the methods for the WalletService
type WalletServiceInterface interface {
	GetBalance(userID int) (float64, error)
	UserExists(userID int) (bool, error)
	Deposit(userID int, amount float64) (*models.Wallet, error)
	Withdraw(userID int, amount float64) (*models.Wallet, error)
	Transfer(fromUserID, toUserID int, amount float64) (*models.Wallet, error)
}

// WalletService handles wallet operations using the repository interface
type WalletService struct {
	walletRepo         WalletRepositoryInterface
	transactionService transaction.TransactionServiceInterface
}

// NewWalletService creates a new WalletService with the provided repository
func NewWalletService(walletRepo WalletRepositoryInterface, transactionService transaction.TransactionServiceInterface) *WalletService {
	return &WalletService{walletRepo: walletRepo, transactionService: transactionService}
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
	err = ws.transactionService.RecordTransaction(nil, &userID, "deposit", amount)
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
	err = ws.transactionService.RecordTransaction(&userID, nil, "withdraw", amount)
	if err != nil {
		return nil, err
	}

	// Return the updated wallet (including balance and updated_at)
	return wallet, nil
}

// Transfer subtracts from one user and adds to another, returning the updated Wallet for the from_user
func (ws *WalletService) Transfer(fromUserID, toUserID int, amount float64) (*models.Wallet, error) {
	// Withdraw from the sender's wallet and get the updated wallet
	wallet, err := ws.walletRepo.Transfer(fromUserID, toUserID, amount)
	if err != nil {
		return wallet, err
	}

	// Record the transfer transaction
	err = ws.transactionService.RecordTransaction(&fromUserID, &toUserID, "transfer", amount)
	if err != nil {
		return wallet, err
	}

	return wallet, nil
}
