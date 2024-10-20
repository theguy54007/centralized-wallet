package user

import (
	"centralized-wallet/internal/models"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type UserServiceInterface interface {
	RegisterUser(email, password string) (*models.User, error)
	LoginUser(email, password string) (*models.User, error)
}

type UserService struct {
	repo UserRepositoryInterface
}

func NewUserService(repo UserRepositoryInterface) *UserService {
	return &UserService{
		repo: repo,
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

	// Step 1: Create the user within the transaction
	user, err := us.repo.CreateUser(email, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
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
	err = verifyPassword(user.Password, password)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	return &models.User{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

// VerifyPassword compares the hashed password with the plain text password
func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
