package repository

import (
	"database/sql"
	"fmt"
)

// WalletRepositoryInterface defines the methods for wallet operations
type WalletRepositoryInterface interface {
	GetWalletBalance(userID int) (float64, error)
	Deposit(userID int, amount float64) error
	Withdraw(userID int, amount float64) error
	Transfer(fromUserID int, toUserID int, amount float64) error
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
func (repo *WalletRepository) Deposit(userID int, amount float64) error {
	query := "UPDATE wallets SET balance = balance + $1 WHERE user_id = $2"
	_, err := repo.db.Exec(query, amount, userID)
	return err
}

// Withdraw subtracts an amount from the wallet balance
func (repo *WalletRepository) Withdraw(userID int, amount float64) error {
	// Check if the user has enough balance
	var balance float64
	err := repo.db.QueryRow("SELECT balance FROM wallets WHERE user_id = $1", userID).Scan(&balance)
	if err != nil {
		return err
	}
	if balance < amount {
		return fmt.Errorf("insufficient funds")
	}

	// Withdraw the amount
	query := "UPDATE wallets SET balance = balance - $1 WHERE user_id = $2"
	_, err = repo.db.Exec(query, amount, userID)
	return err
}

// Transfer sends money from one user to another
func (repo *WalletRepository) Transfer(fromUserID int, toUserID int, amount float64) error {
	// Start a transaction to ensure both operations succeed
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}

	// Check if fromUserID exists
	fromUserExists, err := repo.UserExists(fromUserID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !fromUserExists {
		tx.Rollback()
		return fmt.Errorf("from_user_id does not exist")
	}

	// Check if toUserID exists
	toUserExists, err := repo.UserExists(toUserID)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !toUserExists {
		tx.Rollback()
		return fmt.Errorf("to_user_id does not exist")
	}

	// Withdraw from the sender's wallet
	if err := repo.Withdraw(fromUserID, amount); err != nil {
		tx.Rollback()
		return err
	}

	// Deposit into the recipient's wallet
	if err := repo.Deposit(toUserID, amount); err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	return tx.Commit()
}
