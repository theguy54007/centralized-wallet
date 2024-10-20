package wallet

import (
	"centralized-wallet/internal/models"
	"database/sql"
	"fmt"
)

// WalletRepositoryInterface defines the methods for wallet operations
type WalletRepositoryInterface interface {
	GetWalletBalance(userID int) (float64, error)
	Deposit(userID int, amount float64) (*models.Wallet, error)
	Withdraw(userID int, amount float64) (*models.Wallet, error)
	Transfer(fromUserID int, toUserID int, amount float64) (*models.Wallet, error)
	UserExists(userID int) (bool, error)
}

type WalletRepository struct {
	db *sql.DB
}

// Ensure WalletRepository implements WalletRepositoryInterface
var _ WalletRepositoryInterface = &WalletRepository{}

// NewWalletRepository creates a new instance of WalletRepository
func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// Check if a user exists
func (repo *WalletRepository) UserExists(userID int) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM wallets WHERE user_id = $1)"
	err := repo.db.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetWalletBalance retrieves the balance for a given user ID
func (repo *WalletRepository) GetWalletBalance(userID int) (float64, error) {
	var balance float64
	query := "SELECT balance FROM wallets WHERE user_id = $1"
	err := repo.db.QueryRow(query, userID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("no wallet found for user: %d", userID)
		}
		return 0, err
	}
	return balance, nil
}

// Deposit adds an amount to the wallet balance
// Deposit updates the user's balance and returns the updated Wallet struct
func (repo *WalletRepository) Deposit(userID int, amount float64) (*models.Wallet, error) {
	query := "UPDATE wallets SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2 RETURNING id, user_id, balance, created_at, updated_at"
	row := repo.db.QueryRow(query, amount, userID)

	// Create a Wallet struct to store the result
	var wallet models.Wallet
	err := row.Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// Withdraw deducts the amount from the user's wallet and returns the updated balance and updated_at time
func (repo *WalletRepository) Withdraw(userID int, amount float64) (*models.Wallet, error) {
	// Check if the user has enough balance
	var balance float64
	err := repo.db.QueryRow("SELECT balance FROM wallets WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		return nil, err
	}
	if balance < amount {
		return nil, fmt.Errorf("insufficient funds")
	}

	// Withdraw the amount
	query := "UPDATE wallets SET balance = balance - $1, updated_at = NOW() WHERE user_id = $2 RETURNING balance, updated_at"
	row := repo.db.QueryRow(query, amount, userID)

	// Fetch updated balance and updated_at
	var updatedWallet models.Wallet
	err = row.Scan(&updatedWallet.Balance, &updatedWallet.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Return the updated wallet information
	return &updatedWallet, nil
}

// Transfer sends money from one user to another and returns the updated Wallet struct for the from_user
func (repo *WalletRepository) Transfer(fromUserID int, toUserID int, amount float64) (wallet *models.Wallet, err error) {
	// Start a transaction to ensure both operations succeed
	tx, err := repo.db.Begin()
	if err != nil {
		return wallet, err
	}

	// Check if fromUserID exists
	fromUserExists, err := repo.UserExists(fromUserID)
	if err != nil {
		tx.Rollback()
		return wallet, err
	}
	if !fromUserExists {
		tx.Rollback()
		return wallet, fmt.Errorf("from_user_id does not exist")
	}

	// Check if toUserID exists
	toUserExists, err := repo.UserExists(toUserID)
	if err != nil {
		tx.Rollback()
		return wallet, err
	}
	if !toUserExists {
		tx.Rollback()
		return wallet, fmt.Errorf("to_user_id does not exist")
	}

	// Withdraw from the sender's wallet and get the updated wallet struct for the sender
	wallet, err = repo.Withdraw(fromUserID, amount)
	if err != nil {
		tx.Rollback()
		return wallet, err
	}

	// Deposit into the recipient's wallet (no need to return the wallet for the recipient)
	if _, err := repo.Deposit(toUserID, amount); err != nil {
		tx.Rollback()
		return wallet, err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return wallet, err
	}

	// Return the updated wallet for the sender (already fetched from the Withdraw method)
	return wallet, nil
}
