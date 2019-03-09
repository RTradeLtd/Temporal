package v3

import (
	"context"

	"go.uber.org/zap"

	"github.com/RTradeLtd/Temporal/api/v3/proto/auth"
)

// AuthService implements TemporalAuthService
type AuthService struct {
	l *zap.SugaredLogger
}

// Register returns the Temporal API status
func (a *AuthService) Register(ctx context.Context, req *auth.RegisterReq) (*auth.User, error) {
	return nil, nil
}

// Recover facilitates account recovery
func (a *AuthService) Recover(ctx context.Context, req *auth.RecoverReq) (*auth.User, error) {
	return nil, nil
}

// Login accepts credentials and returns a token for use with further requests.
func (a *AuthService) Login(ctx context.Context, req *auth.Credentials) (*auth.Token, error) {
	return nil, nil
}

// Account returns the account associated with an authenticated request.
func (a *AuthService) Account(ctx context.Context, req *auth.Empty) (*auth.User, error) {
	return nil, nil
}

// Update facilitates modification of the account associated with an
// authenticated request.
func (a *AuthService) Update(ctx context.Context, req *auth.UpdateReq) (*auth.User, error) {
	return nil, nil
}

// Refresh provides a refreshed token associated with an authenticated request.
func (a *AuthService) Refresh(ctx context.Context, req *auth.Empty) (*auth.Token, error) {
	return nil, nil
}
