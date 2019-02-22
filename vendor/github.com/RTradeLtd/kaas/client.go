package kaas

import (
	"fmt"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/grpc/dialer"
	pb "github.com/RTradeLtd/grpc/krab"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client is how we interface with the kaas grpc key manager
type Client struct {
	pb.ServiceClient
	conn *grpc.ClientConn
}

// NewClient is used to instantiate our kaas client in primary or fallback mode
func NewClient(opts config.Services, fallback bool) (*Client, error) {
	var (
		dialOpts   []grpc.DialOption
		krabConfig config.Krab
	)
	if fallback {
		krabConfig = opts.KrabFallback
	} else {
		krabConfig = opts.Krab
	}
	if krabConfig.TLS.CertPath != "" {
		creds, err := credentials.NewClientTLSFromFile(krabConfig.TLS.CertPath, "")
		if err != nil {
			return nil, fmt.Errorf("could not load tls cert: %s", err)
		}
		dialOpts = append(dialOpts,
			grpc.WithTransportCredentials(creds),
			grpc.WithPerRPCCredentials(dialer.NewCredentials(krabConfig.AuthKey, true)))
	} else {
		dialOpts = append(dialOpts,
			grpc.WithInsecure(),
			grpc.WithPerRPCCredentials(dialer.NewCredentials(krabConfig.AuthKey, false)))
	}
	conn, err := grpc.Dial(krabConfig.URL, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:          conn,
		ServiceClient: pb.NewServiceClient(conn),
	}, nil
}

// Close shuts down the client's gRPC connection
func (kc *Client) Close() error { return kc.conn.Close() }
