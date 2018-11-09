package grpc

import (
	"context"
	"fmt"

	"github.com/RTradeLtd/config"
	pb "github.com/RTradeLtd/grpc/temporal"
	"github.com/RTradeLtd/grpc/temporal/request"
	"github.com/RTradeLtd/grpc/temporal/response"
	"google.golang.org/grpc"
)

// Client is how we interface with the GRPC server as a client
type Client struct {
	GC *grpc.ClientConn
	SC pb.SignerClient
}

// NewGRPCClient generates our GRPC API client
func NewGRPCClient(cfg *config.TemporalConfig, insecure bool) (*Client, error) {
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
	return &Client{
		GC: gconn,
		SC: sconn,
	}, nil
}

// GetSignedMessage is used to return a signed a message from our GRPC API Server
func (c *Client) GetSignedMessage(ctx context.Context, req *request.SignRequest) (*response.SignResponse, error) {
	sconn := pb.NewSignerClient(c.GC)
	c.SC = sconn
	return c.SC.GetSignedMessage(ctx, req)
}
