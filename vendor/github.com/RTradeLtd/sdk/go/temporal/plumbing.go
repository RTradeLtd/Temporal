package temporal

import (
	"context"
	"crypto/tls"

	"github.com/RTradeLtd/sdk/go/temporal/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ConnOpts denotes options for connecting with the Temporal API
type ConnOpts struct {
	Advanced *AdvancedConnOpts
}

// AdvancedConnOpts denote advanced configuration options
type AdvancedConnOpts struct {
	API string      // leave blank for default
	TLS *tls.Config // leave nil to disable
}

// Connect creates a new authenticated connection to the Temporal API
func Connect(ctx context.Context, token *auth.Token, opts ConnOpts) (*grpc.ClientConn, error) {
	dialOpts := make([]grpc.DialOption, 0)
	if token != nil {
		dialOpts = []grpc.DialOption{
			grpc.WithPerRPCCredentials(NewCredentials(token, true, opts)),
		}
	}

	var addr = apiAddress
	if opts.Advanced != nil {
		if opts.Advanced.API != "" {
			addr = opts.Advanced.API
		}
		if opts.Advanced.TLS == nil {
			dialOpts = append(dialOpts, grpc.WithInsecure())
		} else {
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(
				credentials.NewTLS(opts.Advanced.TLS),
			))
		}
	}

	return grpc.DialContext(ctx, addr, dialOpts...)
}

// Authenticate creates a new authenticated connection to the Temporal API
func Authenticate(ctx context.Context, creds *auth.Credentials, opts ConnOpts) (*grpc.ClientConn, error) {
	conn, err := Connect(ctx, nil, opts)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	tok, err := auth.NewTemporalAuthClient(conn).Login(ctx, creds)
	if err != nil {
		return nil, err
	}
	return Connect(ctx, tok, opts)
}
