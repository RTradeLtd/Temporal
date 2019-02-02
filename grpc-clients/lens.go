package clients

import (
	"context"
	"fmt"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/grpc/dialer"
	pb "github.com/RTradeLtd/grpc/lens"
	"github.com/RTradeLtd/grpc/lens/request"
	"github.com/RTradeLtd/grpc/lens/response"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultURL = "127.0.0.1:9998"
)

// LensClient is a lens client used to make requests to the Lens gRPC server
type LensClient struct {
	conn *grpc.ClientConn
	pb.IndexerAPIClient
}

// IndexerAPIClient is the interface used by our grpc lens client
type IndexerAPIClient interface {
	// Index is used to submit content to be indexed by the lens system
	Index(ctx context.Context, in *request.Index, opts ...grpc.CallOption) (*response.Index, error)
	// Search is used to perform a configurable search against the Lens index
	Search(ctx context.Context, in *request.Search, opts ...grpc.CallOption) (*response.Results, error)
}

// NewLensClient is used to generate our lens client
func NewLensClient(opts config.Services) (*LensClient, error) {
	dialOpts := make([]grpc.DialOption, 0)
	if opts.Lens.TLS.CertPath != "" {
		creds, err := credentials.NewClientTLSFromFile(opts.Lens.TLS.CertPath, "")
		if err != nil {
			return nil, fmt.Errorf("could not load tls cert: %s", err)
		}
		dialOpts = append(dialOpts,
			grpc.WithTransportCredentials(creds),
			grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Lens.AuthKey, true)))
	} else {
		dialOpts = append(dialOpts,
			grpc.WithInsecure(),
			grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Lens.AuthKey, false)))
	}
	var url string
	if opts.Lens.URL == "" {
		url = defaultURL
	} else {
		url = opts.Lens.URL
	}

	conn, err := grpc.Dial(url, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &LensClient{
		conn:             conn,
		IndexerAPIClient: pb.NewIndexerAPIClient(conn),
	}, nil
}

// Close shuts down the client's gRPC connection
func (l *LensClient) Close() { l.conn.Close() }
