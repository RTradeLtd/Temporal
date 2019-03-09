package v3

import (
	"context"
	"net"

	"github.com/RTradeLtd/Temporal/api/v3/proto/ipfs"

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
	ipfs  ipfs.TemporalIPFSServer

	l *zap.SugaredLogger
}

// New initializes a new V3 service from concrete implementations of the
// V3 subservices
func New(
	l *zap.SugaredLogger,
	coreService *CoreService,
	authService *AuthService,
	storeService *StoreService,
	ipfsService *IPFSService,
) *V3 {
	return &V3{
		core:  coreService,
		auth:  authService,
		store: storeService,
		ipfs:  ipfsService,

		l: l,
	}
}

// Run spins up daemon server
func (v *V3) Run(ctx context.Context, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		v.l.Errorw("failed to listen on given address", "address", address)
		return err
	}

	// initialize server
	v.l.Debug("registering services")
	var server = grpc.NewServer()
	core.RegisterTemporalCoreServer(server, v.core)
	auth.RegisterTemporalAuthServer(server, v.auth)
	store.RegisterTemporalStoreServer(server, v.store)
	ipfs.RegisterTemporalIPFSServer(server, v.ipfs)
	v.l.Debug("services registered")

	// interrupt server gracefully if context is cancelled
	go func() {
		for {
			select {
			case <-ctx.Done():
				v.l.Info("shutting down server")
				server.GracefulStop()
				return
			}
		}
	}()

	// spin up server
	v.l.Info("spinning up server")
	return server.Serve(listener)
}
