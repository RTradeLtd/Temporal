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

// Status returns the Temporal API status
func (a *AuthService) Status(context.Context, *auth.Message) (*auth.Message, error) {
	return &auth.Message{
		Message: "hello world",
	}, nil
}
