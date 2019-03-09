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
func (c *CoreService) Status(ctx context.Context, req *core.Empty) (*core.ServiceStatus, error) {
	return nil, nil
}

// Statistics returns statistics about this Temporal instance
func (c *CoreService) Statistics(ctx context.Context, req *core.Empty) (*core.ServiceStatistics, error) {
	return nil, nil
}
