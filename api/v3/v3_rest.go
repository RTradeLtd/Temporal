package v3

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/RTradeLtd/Temporal/api/v3/proto/auth"
	"github.com/RTradeLtd/Temporal/api/v3/proto/core"
	"github.com/RTradeLtd/Temporal/api/v3/proto/ipfs"
	"github.com/RTradeLtd/Temporal/api/v3/proto/store"
)

// RESTGatewayOptions denotes options for the RESTful gRPC gateway
type RESTGatewayOptions struct {
	Address string

	DialAddress string
	DialOptions []grpc.DialOption
}

// REST runs the RESTful reverse proxy for the Temporal V3 gRPC API
func REST(ctx context.Context, opts RESTGatewayOptions) error {
	var mux = runtime.NewServeMux(
		runtime.WithMetadata(annotator))

	// register all gateways
	if err := core.RegisterTemporalCoreHandlerFromEndpoint(
		ctx,
		mux,
		opts.DialAddress,
		opts.DialOptions); err != nil {
		return err
	}
	if err := auth.RegisterTemporalAuthHandlerFromEndpoint(
		ctx,
		mux,
		opts.DialAddress,
		opts.DialOptions); err != nil {
		return err
	}
	if err := ipfs.RegisterTemporalIPFSHandlerFromEndpoint(
		ctx,
		mux,
		opts.DialAddress,
		opts.DialOptions); err != nil {
		return err
	}
	if err := store.RegisterTemporalStoreHandlerFromEndpoint(
		ctx,
		mux,
		opts.DialAddress,
		opts.DialOptions); err != nil {
		return err
	}

	// spin up server
	return http.ListenAndServe(opts.Address, mux)
}

// annotator is used by gateway to read values from HTTP requests into gRPC
// metadata
func annotator(ctx context.Context, req *http.Request) metadata.MD {
	return metadata.MD{
		"authorization": []string{req.Header.Get("Authorization")},
	}
}
