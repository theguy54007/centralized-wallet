package wallet

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/utils"
	"database/sql"
)

// WalletRepositoryInterface defines the methods for wallet operations
type WalletRepositoryInterface interface {
	Begin() (*sql.Tx, error)
	CreateWallet(wallet *models.Wallet) error // Removed transaction
	IsWalletNumberExists(walletNumber string) (bool, error)
	GetWalletBalance(userID int) (float64, error)
	GetWalletByUserID(userID int) (*models.Wallet, error)
	Deposit(tx *sql.Tx, userID int, amount float64) (*models.Wallet, error)
	Withdraw(tx *sql.Tx, userID int, amount float64) (*models.Wallet, error)
	UserExists(userID int) (bool, error)
	FindByWalletNumber(walletNumber string) (*models.Wallet, error)
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

// Begin a transaction
func (repo *WalletRepository) Begin() (*sql.Tx, error) {
	return repo.db.Begin()
}

func (repo *WalletRepository) CreateWallet(wallet *models.Wallet) error {
	query := `INSERT INTO wallets (user_id, balance, wallet_number, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5)`
	_, err := repo.db.Exec(query, wallet.UserID, wallet.Balance, wallet.WalletNumber, wallet.CreatedAt, wallet.UpdatedAt)
	return err
}

func (repo *WalletRepository) GetWalletByUserID(userID int) (*models.Wallet, error) {
	var wallet models.Wallet
	query := "SELECT id, user_id, wallet_number, balance, created_at, updated_at FROM wallets WHERE user_id = $1"
	err := repo.db.QueryRow(query, userID).Scan(&wallet.ID, &wallet.UserID, &wallet.WalletNumber, &wallet.Balance, &wallet.CreatedAt, &wallet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &wallet, nil
}

// GetWalletBalance retrieves the balance for a given user ID
func (repo *WalletRepository) GetWalletBalance(userID int) (float64, error) {
	var balance float64
	query := "SELECT balance FROM wallets WHERE user_id = $1"
	err := repo.db.QueryRow(query, userID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, utils.RepoErrWalletNotFound
		}
		return 0, err
	}
	return balance, nil
}

// Deposit adds an amount to the wallet balance
// Deposit updates the user's balance and returns the updated Wallet struct
func (repo *WalletRepository) Deposit(tx *sql.Tx, userID int, amount float64) (*models.Wallet, error) {
	query := "UPDATE wallets SET balance = balance + $1, updated_at = NOW() WHERE user_id = $2 RETURNING id, user_id, balance, wallet_number, created_at, updated_at"
	row := tx.QueryRow(query, amount, userID)

	// Create a Wallet struct to store the result
	var updatedWallet models.Wallet
	err := row.Scan(&updatedWallet.ID, &updatedWallet.UserID, &updatedWallet.Balance, &updatedWallet.WalletNumber, &updatedWallet.CreatedAt, &updatedWallet.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &updatedWallet, nil
}

// Withdraw deducts the amount from the user's wallet and returns the updated balance and updated_at time
func (repo *WalletRepository) Withdraw(tx *sql.Tx, userID int, amount float64) (*models.Wallet, error) {
	// Withdraw the amount
	query := "UPDATE wallets SET balance = balance - $1, updated_at = NOW() WHERE user_id = $2 RETURNING balance, wallet_number, updated_at"
	// row := repo.db.QueryRow(query, amount, userID)
	row := tx.QueryRow(query, amount, userID)

	// Fetch updated balance and updated_at
	var updatedWallet models.Wallet
	err := row.Scan(&updatedWallet.Balance, &updatedWallet.WalletNumber, &updatedWallet.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Return the updated wallet information
	return &updatedWallet, nil
}

func (repo *WalletRepository) FindByWalletNumber(walletNumber string) (*models.Wallet, error) {
	query := "SELECT id, user_id, balance, wallet_number, updated_at FROM wallets WHERE wallet_number = $1"
	wallet := &models.Wallet{}
	err := repo.db.QueryRow(query, walletNumber).Scan(&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.WalletNumber, &wallet.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (repo *WalletRepository) IsWalletNumberExists(walletNumber string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM wallets WHERE wallet_number = $1)"
	err := repo.db.QueryRow(query, walletNumber).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
