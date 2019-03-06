package v3

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"github.com/RTradeLtd/Temporal/api/v3/proto/core"
)

// RESTGatewayOptions denotes options for the RESTful gRPC gateway
type RESTGatewayOptions struct {
	Address string

	DialAddress string
	DialOptions []grpc.DialOption
}

// REST runs the RESTful reverse proxy for the Temporal V3 gRPC API
func REST(ctx context.Context, opts RESTGatewayOptions) error {
	var mux = runtime.NewServeMux()

	if err := core.RegisterTemporalCoreHandlerFromEndpoint(
		ctx,
		mux,
		opts.DialAddress,
		opts.DialOptions); err != nil {
		return err
	}

	return http.ListenAndServe(opts.Address, mux)
}
