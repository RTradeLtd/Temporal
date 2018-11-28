package kaas

import (
	"fmt"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/grpc/dialer"
	pb "github.com/RTradeLtd/grpc/krab"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultURL = "127.0.0.1:10000"
)

// Client is our connection to the grpc server
type Client struct {
	conn *grpc.ClientConn
	pb.ServiceClient
}

// NewClient is used to generate our lens client
func NewClient(opts config.Endpoints) (*Client, error) {
	dialOpts := make([]grpc.DialOption, 0)
	if opts.Krab.TLS.CertPath != "" {
		creds, err := credentials.NewClientTLSFromFile(opts.Krab.TLS.CertPath, "")
		if err != nil {
			return nil, fmt.Errorf("could not load tls cert: %s", err)
		}
		dialOpts = append(dialOpts,
			grpc.WithTransportCredentials(creds),
			grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Krab.AuthKey, true)))
	} else {
		dialOpts = append(dialOpts,
			grpc.WithInsecure(),
			grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Krab.AuthKey, false)))
	}
	fmt.Println(opts.Krab.URL)
	var url string
	if opts.Krab.URL == "" {
		url = defaultURL
	} else {
		url = opts.Krab.URL
	}

	conn, err := grpc.Dial(url, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		conn:          conn,
		ServiceClient: pb.NewServiceClient(conn),
	}, nil
}

// Close shuts down the client's gRPC connection
func (c *Client) Close() error { return c.conn.Close() }
