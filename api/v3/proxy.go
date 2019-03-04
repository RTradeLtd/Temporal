package v3

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"github.com/RTradeLtd/Temporal/api/v3/proto/core"
)

// REST runs the RESTful reverse proxy for the Temporal V3 gRPC API
func REST(serviceAddress string) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := core.RegisterTemporalCoreHandlerFromEndpoint(ctx, mux, serviceAddress, opts)
	if err != nil {
		return err
	}

	return http.ListenAndServe(":8080", mux)
}
