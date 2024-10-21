package utils

import "errors"

type AppError struct {
	Code    int
	Message string
	Err     error
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Message == t.Message
}

// Define all custom errors here
var (

	//  Error with HTTP status code mappings
	// 400 level errors
	ErrUnauthorized        = NewAppError(401, "Authorization token is required", nil)
	ErrInvalidToken        = NewAppError(401, "Invalid token", nil)
	ErrTokenExpired        = NewAppError(401, "Token expired", nil)
	ErrInvalidUserId       = NewAppError(401, "Invalid user ID", nil)
	ErrInvalidEmailFormat  = NewAppError(400, "Invalid email format", nil)
	ErrPasswordTooShort    = NewAppError(400, "Password must be at least 6 characters", nil)
	ErrInvalidCredentials  = NewAppError(401, "Invalid email or password", nil)
	ErrInvalidRequest      = NewAppError(400, "Invalid request data", nil)
	ErrEmailAlreadyInUse   = NewAppError(400, "Email already in use", nil)
	ErrUserNotFound        = NewAppError(400, "User not found", nil)
	ErrWalletAlreadyExists = NewAppError(400, "wallet already exists for this user", nil)
	ErrWalletNotFound      = NewAppError(400, "Wallet not found", nil)
	ErrorWalletNumber      = NewAppError(400, "Invalid wallet number", nil)
	ErrorInvalidOrder      = NewAppError(400, "Invalid order, must be 'asc' or 'desc'", nil)
	ErrorInvalidLimit      = NewAppError(400, "Invalid limit, must be between 1 and 100", nil)
	ErrorInvalidOffset     = NewAppError(400, "Invalid offset, must be a non-negative integer", nil)
	ErrorInsufficientFunds = NewAppError(400, "Insufficient funds", nil)

	// 500 level errors
	ErrInternalServerError   = NewAppError(500, "Internal server error", nil)
	ErrDatabaseError         = NewAppError(500, "Database operation failed", nil)
	ErrUserCreationFailed    = NewAppError(500, "Could not create user", nil)
	ErrTokenGenerationFailed = NewAppError(500, "Could not generate token", nil)

	// Repository errors
	RepoErrWalletNotFound    = errors.New("from_wallet_number does not exist")
	RepoErrUserNotFound      = errors.New("from_user does not exist")
	RepoErrToUserNotFound    = errors.New("to_user does not exist")
	RepoErrToWalletNotFound  = errors.New("to_wallet_number does not exist")
	RepoErrInsufficientFunds = errors.New("insufficient funds")
	RepoErrDatabaseOperation = errors.New("database operation failed")
	RepoErrTransactionFailed = errors.New("transaction failed")

	// Service errors
	ServiceErrWalletAlreadyExists = errors.New("wallet already exists for this user")
	ServiceErrWalletNumberNil     = errors.New("either fromWalletNumber or toWalletNumber must be provided")
)
