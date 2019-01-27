package clients

import (
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/grpc/dialer"
	pb "github.com/RTradeLtd/grpc/pay"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultSignerURL = "127.0.0.1:9090"
)

// SignerClient is how we interface with the Signer server as a client
type SignerClient struct {
	pb.SignerClient
	conn *grpc.ClientConn
}

// NewSignerClient is used to instantiate our connection with our grpc payment api
func NewSignerClient(cfg *config.TemporalConfig) (*SignerClient, error) {
	dialOpts := make([]grpc.DialOption, 0)
	if cfg.Pay.TLS.CertPath != "" {
		creds, err := credentials.NewClientTLSFromFile(cfg.Pay.TLS.CertPath, "")
		if err != nil {
			return nil, err
		}
		dialOpts = append(dialOpts,
			grpc.WithTransportCredentials(creds),
			grpc.WithPerRPCCredentials(dialer.NewCredentials(cfg.Pay.AuthKey, true)))
	} else {
		dialOpts = append(dialOpts,
			grpc.WithInsecure(),
			grpc.WithPerRPCCredentials(dialer.NewCredentials(cfg.Pay.AuthKey, false)))
	}
	var url string
	if cfg.Pay.Address == "" {
		url = defaultSignerURL
	} else {
		url = cfg.Pay.Address + ":" + cfg.Pay.Port
	}
	conn, err := grpc.Dial(url, dialOpts...)
	if err != nil {
		return nil, err
	}
	return &SignerClient{
		conn:         conn,
		SignerClient: pb.NewSignerClient(conn),
	}, nil
}

// Close shuts down the client's gRPC connection
func (s *SignerClient) Close() { s.conn.Close() }
