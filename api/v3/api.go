package v3

import (
	"context"
	"net"

	"google.golang.org/grpc"

	"github.com/RTradeLtd/Temporal/api/v3/proto/core"
)

// API is the core Temporal V3 API
type API struct{}

// New initializes a new Daemon
func New() *API {
	return &API{}
}

// Run spins up daemon server
func (a *API) Run(ctx context.Context, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	// initialize server
	server := grpc.NewServer()
	core.RegisterTemporalCoreServer(server, a)

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

// Status returns the Temporal API status
func (a *API) Status(context.Context, *core.Message) (*core.Message, error) {
	return &core.Message{
		Message: "hello world",
	}, nil
}
