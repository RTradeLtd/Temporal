package v3

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/RTradeLtd/Temporal/api/v3/proto/auth"
	"github.com/RTradeLtd/Temporal/api/v3/proto/core"
	"github.com/RTradeLtd/Temporal/api/v3/proto/ipfs"
	"github.com/RTradeLtd/Temporal/api/v3/proto/store"
)

// V3 is the container around Temporal's V3 API (core, auth, and store)
type V3 struct {
	core  core.TemporalCoreServer
	auth  auth.TemporalAuthServer
	store store.TemporalStoreServer
	ipfs  ipfs.TemporalIPFSServer

	verify http.HandlerFunc
	http   *http.Server

	options []grpc.ServerOption

	l *zap.SugaredLogger
}

// Options denotes configuration for the V3 API
type Options struct {
	TLS *tls.Config
}

// New initializes a new V3 service from concrete implementations of the
// V3 subservices
func New(
	l *zap.SugaredLogger,

	coreService *CoreService,
	authService *AuthService,
	storeService *StoreService,
	ipfsService *IPFSService,

	opts Options,
) *V3 {
	var grpcLogger = l.Desugar().Named("grpc")
	grpc_zap.ReplaceGrpcLogger(grpcLogger)

	// set up middleware chain
	var (
		unaryAuth, streamAuth = authService.newAuthInterceptors(
			"/auth.TemporalAuth/Register",
			"/auth.TemporalAuth/Login",
			"/auth.TemporalAuth/Recover")

		zapOpts = []grpc_zap.Option{
			grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
				return zap.Duration("grpc.duration", duration)
			}),
		}

		serverOpts = []grpc.ServerOption{
			grpc_middleware.WithUnaryServerChain(
				unaryAuth,
				grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_zap.UnaryServerInterceptor(grpcLogger, zapOpts...)),
			grpc_middleware.WithStreamServerChain(
				streamAuth,
				grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
				grpc_zap.StreamServerInterceptor(grpcLogger, zapOpts...)),
		}
	)

	// set up TLS if configuration is provided for it
	if opts.TLS != nil {
		l.Infow("setting up TLS")
		serverOpts = append(serverOpts, grpc.Creds(credentials.NewTLS(opts.TLS)))
	} else {
		l.Warn("no TLS configuration found")
	}

	return &V3{
		core:  coreService,
		auth:  authService,
		store: storeService,
		ipfs:  ipfsService,

		verify: authService.httpVerificationHandler,
		http: &http.Server{
			TLSConfig: opts.TLS,
		},

		options: serverOpts,

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
	var server = grpc.NewServer(v.options...)
	core.RegisterTemporalCoreServer(server, v.core)
	auth.RegisterTemporalAuthServer(server, v.auth)
	store.RegisterTemporalStoreServer(server, v.store)
	ipfs.RegisterTemporalIPFSServer(server, v.ipfs)
	v.l.Debug("services registered")

	// set up rest endpoint
	v.l.Debug("setting up REST endpoints")
	var m = chi.NewMux()
	m.Get("/v3/auth/verify", v.verify)
	v.http.Handler = m
	v.http.Addr = address
	v.l.Debug("rest endpoints set up")

	// interrupt server gracefully if context is cancelled
	go func() {
		for {
			select {
			case <-ctx.Done():
				v.l.Info("shutting down server")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				v.http.Shutdown(ctx)
				server.GracefulStop()
				cancel()
				return
			}
		}
	}()

	// spin up server
	v.l.Info("spinning up http server")
	go v.http.ListenAndServe()

	v.l.Info("spinning up grpc services")
	return server.Serve(listener)
}
