package wallet

import (
	"centralized-wallet/internal/repository"
	"centralized-wallet/internal/transaction"
)

// WalletService handles wallet operations using the repository interface
type WalletService struct {
	walletRepo         repository.WalletRepositoryInterface
	transactionService transaction.TransactionServiceInterface
}

// NewWalletService creates a new WalletService with the provided repository
func NewWalletService(walletRepo repository.WalletRepositoryInterface, transactionService transaction.TransactionServiceInterface) *WalletService {
	return &WalletService{walletRepo: walletRepo, transactionService: transactionService}
}

// GetBalance retrieves the balance of a user
func (ws *WalletService) GetBalance(userID int) (float64, error) {
	return ws.walletRepo.GetWalletBalance(userID)
}

// Deposit adds money to the user's wallet and records the transaction
func (ws *WalletService) Deposit(userID int, amount float64) error {
	err := ws.walletRepo.Deposit(userID, amount)
	if err != nil {
		return err
	}
	// Record the deposit transaction
	return ws.transactionService.RecordTransaction(userID, "deposit", amount)
}

// Withdraw subtracts money from the user's wallet and records the transaction
func (ws *WalletService) Withdraw(userID int, amount float64) error {
	err := ws.walletRepo.Withdraw(userID, amount)
	if err != nil {
		return err
	}
	// Record the withdrawal transaction
	return ws.transactionService.RecordTransaction(userID, "withdraw", amount)
}

func (ws *WalletService) Transfer(fromUserID, toUserID int, amount float64) error {
	// Withdraw from the sender's wallet
	err := ws.walletRepo.Withdraw(fromUserID, amount)
	if err != nil {
		return err
	}

	// Record the withdrawal transaction for the sender
	err = ws.transactionService.RecordTransaction(fromUserID, "transfer out", amount)
	if err != nil {
		return err
	}

	// Deposit into the recipient's wallet
	err = ws.walletRepo.Deposit(toUserID, amount)
	if err != nil {
		return err
	}

	// Record the deposit transaction for the recipient
	return ws.transactionService.RecordTransaction(toUserID, "transfer in", amount)
}
