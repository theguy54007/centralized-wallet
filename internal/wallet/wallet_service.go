package wallet

import "centralized-wallet/internal/repository"

// WalletService handles wallet operations using the repository interface
type WalletService struct {
	repo repository.WalletRepositoryInterface
}

// NewWalletService creates a new WalletService with the provided repository
func NewWalletService(repo repository.WalletRepositoryInterface) *WalletService {
	return &WalletService{repo: repo}
}

// GetBalance retrieves the balance of a user
func (ws *WalletService) GetBalance(userID int) (float64, error) {
	return ws.repo.GetWalletBalance(userID)
}

// Deposit adds an amount to the user's wallet
func (ws *WalletService) Deposit(userID int, amount float64) error {
	return ws.repo.Deposit(userID, amount)
}

// Withdraw subtracts an amount from the user's wallet
func (ws *WalletService) Withdraw(userID int, amount float64) error {
	return ws.repo.Withdraw(userID, amount)
}

// Transfer moves money from one user to another
func (ws *WalletService) Transfer(fromUserID int, toUserID int, amount float64) error {
	return ws.repo.Transfer(fromUserID, toUserID, amount)
}
