package clients

import (
	"fmt"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/grpc/dialer"
	pb "github.com/RTradeLtd/grpc/lensv2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultURL = "127.0.0.1:9998"
)

// LensClient is a lens client used to make requests to the Lens gRPC server
type LensClient struct {
	conn *grpc.ClientConn
	pb.LensV2Client
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
		conn:         conn,
		LensV2Client: pb.NewLensV2Client(conn),
	}, nil
}

// Close shuts down the client's gRPC connection
func (l *LensClient) Close() { l.conn.Close() }
