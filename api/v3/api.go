package v3

import (
	"context"
	"net"

	temporalv3 "github.com/RTradeLtd/Temporal/api/v3/proto"
	"google.golang.org/grpc"
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
	temporalv3.RegisterTemporalServer(server, a)

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
func (a *API) Status(context.Context, *temporalv3.Message) (*temporalv3.Message, error) {
	return &temporalv3.Message{
		Message: "hello world",
	}, nil
}
