package apperrors

import "errors"

// Define all custom errors here
var (
	ErrWalletAlreadyExists = errors.New("wallet already exists for this user")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
)
