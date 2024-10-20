package mock_auth

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
)

type MockBlacklistService struct {
	mock.Mock
}

// BlacklistToken mocks the BlacklistToken function
func (m *MockBlacklistService) BlacklistToken(tokenString string, token *jwt.Token) error {
	args := m.Called(tokenString, token)
	return args.Error(0)
}

// IsTokenBlacklisted mocks the IsTokenBlacklisted function
func (m *MockBlacklistService) IsTokenBlacklisted(tokenString string) (bool, error) {
	args := m.Called(tokenString)
	return args.Bool(0), args.Error(1)
}

// RemoveBlacklistedToken mocks the RemoveBlacklistedToken function
func (m *MockBlacklistService) RemoveBlacklistedToken(ctx context.Context, tokenString string) error {
	args := m.Called(ctx, tokenString)
	return args.Error(0)
}
