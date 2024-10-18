package repository

import (
	"centralized-wallet/internal/models"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
)

type UserRepositoryInterface interface {
	CreateUser(email, password string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
}

// Ensure UserRepository implements the UserRepositoryInterface
var _ UserRepositoryInterface = &UserRepository{}

type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user with a hashed password
func (repo *UserRepository) CreateUser(email, password string) (*models.User, error) {
	// Check if email is already in use
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)"
	err := repo.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already in use")
	}

	// Hash the password
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Insert the new user into the database
	query = "INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id"
	var userID int
	err = repo.db.QueryRow(query, email, hashedPassword).Scan(&userID)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:       userID,
		Email:    email,
		Password: hashedPassword,
	}, nil
}

// GetUserByEmail retrieves a user by their email from the database
func (repo *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	query := "SELECT id, email, password FROM users WHERE email = $1"
	err := repo.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// VerifyPassword compares the hashed password with the plain text password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
