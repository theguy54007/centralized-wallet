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

	// Withdraw from the sender's wallet
	if err := repo.Withdraw(fromUserID, amount); err != nil {
		tx.Rollback() // Rollback if anything goes wrong
		return err
	}

	// Deposit into the recipient's wallet
	if err := repo.Deposit(toUserID, amount); err != nil {
		tx.Rollback() // Rollback if anything goes wrong
		return err
	}

	// Commit the transaction
	return tx.Commit()
}
