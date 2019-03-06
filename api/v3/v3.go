package v3

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/RTradeLtd/Temporal/api/v3/proto/auth"
	"github.com/RTradeLtd/Temporal/api/v3/proto/core"
	"github.com/RTradeLtd/Temporal/api/v3/proto/store"
)

// V3 is the container around Temporal's V3 API (core, auth, and store)
type V3 struct {
	core  core.TemporalCoreServer
	auth  auth.TemporalAuthServer
	store store.TemporalStoreServer

	l *zap.SugaredLogger
}

// New initializes a new V3 service from concrete implementations of the
// V3 subservices
func New(
	l *zap.SugaredLogger,
	coreService *CoreService,
	authService *AuthService,
	storeService *StoreService,
) *V3 {
	return &V3{
		core:  coreService,
		auth:  authService,
		store: storeService,

		l: l,
	}
}

// Run spins up daemon server
func (v *V3) Run(ctx context.Context, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	// initialize server
	var server = grpc.NewServer()
	core.RegisterTemporalCoreServer(server, v.core)
	auth.RegisterTemporalAuthServer(server, v.auth)
	store.RegisterTemporalStoreServer(server, v.store)

	// interrupt server gracefully if context is cancelled
	go func() {
		for {
			select {
			case <-ctx.Done():
				server.GracefulStop()
				return
			}
		}
	}()

	// spin up server
	return server.Serve(listener)
}
