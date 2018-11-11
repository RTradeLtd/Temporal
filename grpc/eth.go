package grpc

import (
	"fmt"

	"github.com/RTradeLtd/config"
	pb "github.com/RTradeLtd/grpc/temporal"
	"google.golang.org/grpc"
)

// SignerClient is how we interface with the Signer server as a client
type SignerClient struct {
	pb.SignerClient
	conn *grpc.ClientConn
}

// NewSignerClient instantiates a new Signerclient
func NewSignerClient(cfg *config.TemporalConfig, insecure bool) (*SignerClient, error) {
	grpcAPI := fmt.Sprintf("%s:%s", cfg.API.Payment.Address, cfg.API.Payment.Port)
	var (
		gconn *grpc.ClientConn
		err   error
	)
	if insecure {
		gconn, err = grpc.Dial(grpcAPI, grpc.WithInsecure())
	}
	if err != nil {
		return nil, err
	}
	sconn := pb.NewSignerClient(gconn)
	return &SignerClient{
		conn:         gconn,
		SignerClient: sconn,
	}, nil
}

// Close shuts down the client's gRPC connection
func (s *SignerClient) Close() { s.conn.Close() }
