package v3

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/RTradeLtd/Temporal/api/v3/proto/core"
)

// CoreService implements TemporalCoreService
type CoreService struct {
	l *zap.SugaredLogger
}

// Status returns the Temporal API status
func (c *CoreService) Status(ctx context.Context, req *core.Empty) (*core.ServiceStatus, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// Statistics returns statistics about this Temporal instance
func (c *CoreService) Statistics(ctx context.Context, req *core.Empty) (*core.ServiceStatistics, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}
