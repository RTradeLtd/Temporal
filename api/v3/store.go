package v3

import (
	"context"

	"go.uber.org/zap"

	"github.com/RTradeLtd/Temporal/api/v3/proto/store"
)

// StoreService implements TemporalStoreService
type StoreService struct {
	l *zap.SugaredLogger
}

// Status returns the Temporal API status
func (s *StoreService) Status(context.Context, *store.Message) (*store.Message, error) {
	return &store.Message{
		Message: "hello world",
	}, nil
}
