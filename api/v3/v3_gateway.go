package v3

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/go-chi/chi"
	"go.uber.org/zap"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/RTradeLtd/Temporal/api/v3/proto/auth"
	"github.com/RTradeLtd/Temporal/api/v3/proto/core"
	"github.com/RTradeLtd/Temporal/api/v3/proto/ipfs"
	"github.com/RTradeLtd/Temporal/api/v3/proto/store"
	"github.com/RTradeLtd/Temporal/log"
)

// RESTGatewayOptions denotes options for the RESTful gRPC gateway
type RESTGatewayOptions struct {
	Address string
	TLS     *tls.Config

	DialAddress string
	DialOptions []grpc.DialOption
}

// REST runs the RESTful reverse proxy for the Temporal V3 gRPC API
func REST(
	ctx context.Context,
	l *zap.SugaredLogger,
	handlers map[string]http.HandleFunc,
	opts RESTGatewayOptions,
) error {
	var gateway = runtime.NewServeMux(
		runtime.WithMetadata(gatewayAnnotator))

	// register all gateways
	if err := core.RegisterTemporalCoreHandlerFromEndpoint(
		ctx, gateway, opts.DialAddress, opts.DialOptions); err != nil {
		return err
	}
	if err := auth.RegisterTemporalAuthHandlerFromEndpoint(
		ctx, gateway, opts.DialAddress, opts.DialOptions); err != nil {
		return err
	}
	if err := ipfs.RegisterTemporalIPFSHandlerFromEndpoint(
		ctx, gateway, opts.DialAddress, opts.DialOptions); err != nil {
		return err
	}
	if err := store.RegisterTemporalStoreHandlerFromEndpoint(
		ctx, gateway, opts.DialAddress, opts.DialOptions); err != nil {
		return err
	}

	// register routes
	mux := chi.NewMux()
	mux.Use(log.NewMiddleware(l.Named("requests")))
	mux.Route("/v3", func(r *chi.Router) {
		r.Handle("/", gateway)
		for path, fn := range handlers {
			r.HandleFunc(path, fn)
		}
	})
	srv := &http.Server{
		Addr: opts.Address,
		TLS:  opts.TLS,

		Handler: mux,
	}

	// spin up server
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	return srv.ListenAndServe()
}

// gatewayAnnotator is used by gateway to read values from HTTP requests into
// gRPC metadata
func gatewayAnnotator(ctx context.Context, req *http.Request) metadata.MD {
	return metadata.MD{
		"authorization": []string{req.Header.Get("Authorization")},
	}
}
