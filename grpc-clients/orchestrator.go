package clients

import (
	"fmt"

	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/grpc/dialer"
	nexus "github.com/RTradeLtd/grpc/nexus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// IPFSOrchestratorClient is a lighweight container for the orchestrator's
// gRPC API client
type IPFSOrchestratorClient struct {
	nexus.ServiceClient
	conn *grpc.ClientConn
}

// NewOcrhestratorClient instantiates a new orchestrator API client
func NewOcrhestratorClient(opts config.Nexus) (*IPFSOrchestratorClient, error) {
	c := &IPFSOrchestratorClient{}
	// set up parameters for core conn
	dialOpts := make([]grpc.DialOption, 0)
	if opts.TLS.CertPath != "" {
		creds, err := credentials.NewClientTLSFromFile(opts.TLS.CertPath, "")
		if err != nil {
			return nil, fmt.Errorf("could not load tls cert: %s", err)
		}
		dialOpts = append(dialOpts,
			grpc.WithTransportCredentials(creds),
			grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Key, true)))
	} else {
		dialOpts = append(dialOpts,
			grpc.WithInsecure(),
			grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Key, false)))
	}

	// connect to orchestrator
	var err error
	c.conn, err = grpc.Dial(opts.Host+":"+opts.Port, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core service: %s", err.Error())
	}
	c.ServiceClient = nexus.NewServiceClient(c.conn)
	return c, nil
}

// Close shuts down the client's gRPC connection
func (i *IPFSOrchestratorClient) Close() { i.conn.Close() }
