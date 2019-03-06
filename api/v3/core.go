package v3

import (
	"context"

	"go.uber.org/zap"

	"github.com/RTradeLtd/Temporal/api/v3/proto/core"
)

// CoreService implements TemporalCoreService
type CoreService struct {
	l *zap.SugaredLogger
}

// Status returns the Temporal API status
func (c *CoreService) Status(context.Context, *core.Message) (*core.Message, error) {
	return &core.Message{
		Message: "hello world",
	}, nil
}
