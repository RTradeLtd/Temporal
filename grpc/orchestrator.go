package grpc

import (
	"fmt"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/grpc/dialer"
	ipfs_orchestrator "github.com/RTradeLtd/ipfs-orchestrator/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// IPFSOrchestratorClient is a lighweight container for the orchestrator's
// gRPC API client
type IPFSOrchestratorClient struct {
	ipfs_orchestrator.ServiceClient
	grpc *grpc.ClientConn
}

// New instantiates a new orchestrator API client
func New(opts config.Orchestrator, devMode bool) (*IPFSOrchestratorClient, error) {
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
			grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Key, !devMode)))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	// connect to orchestrator
	var err error
	c.grpc, err = grpc.Dial(opts.Host+":"+opts.Port, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to core service: %s", err.Error())
	}
	c.ServiceClient = ipfs_orchestrator.NewServiceClient(c.grpc)
	return c, nil
}

// Close shuts down the client's gRPC connection
func (i *IPFSOrchestratorClient) Close() { i.grpc.Close() }
