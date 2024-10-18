package user

import (
	"centralized-wallet/internal/models"
	"centralized-wallet/internal/repository"
	"errors"
)

type UserService struct {
	repo repository.UserRepositoryInterface
}

func NewUserService(repo repository.UserRepositoryInterface) *UserService {
	return &UserService{repo: repo}
}

// Business logic for user registration
func (us *UserService) RegisterUser(email, password string) (*models.User, error) {
	return us.repo.CreateUser(email, password)
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
