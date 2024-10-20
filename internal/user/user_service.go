package user

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/repository"
	"centralized-wallet/internal/wallet"
	"errors"
	"fmt"
)

type UserServiceInterface interface {
	RegisterUser(email, password string) (*models.User, error)
	LoginUser(email, password string) (*models.User, error)
}

type UserService struct {
	repo          repository.UserRepositoryInterface
	walletService wallet.WalletServiceInterface
}

func NewUserService(repo repository.UserRepositoryInterface, walletService *wallet.WalletService) *UserService {
	return &UserService{
		repo:          repo,
		walletService: walletService,
	}
}

// Business logic for user registration
// RegisterUser creates a new user and wallet, ensuring both operations are atomic
func (us *UserService) RegisterUser(email, password string) (*models.User, error) {
	// Check if the email is already in use before starting the transaction
	emailInUse, err := us.repo.IsEmailInUse(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %v", err)
	}
	if emailInUse {
		return nil, errors.New("email already in use")
	}

	// Begin the transaction via the UserRepository
	tx, err := us.repo.BeginTransaction()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Step 1: Create the user within the transaction
	user, err := us.repo.CreateUserWithTx(tx, email, password)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Step 2: Create the wallet for the user within the same transaction
	_, err = us.walletService.CreateWalletWithTx(tx, user.ID)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create wallet: %v", err)
	}

	// Step 3: Commit the transaction if both operations succeed
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return user, nil
}

// Business logic for user login (to be added later)
func (us *UserService) LoginUser(email, password string) (*models.User, error) {
	// Find the user by email
	user, err := us.repo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}

	// Check if user is nil (in case the repository returns a nil user)
	if user == nil {
		return nil, errors.New("invalid email or password")
	}

	// Verify the provided password with the stored hashed password
	err = repository.VerifyPassword(user.Password, password)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	return &models.User{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}
