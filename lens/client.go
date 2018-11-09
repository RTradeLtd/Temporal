package lens

import (
	"fmt"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/grpc/dialer"
	pb "github.com/RTradeLtd/grpc/lens"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultURL = "127.0.0.1:9998"
)

// Client is a lens client used to make requests to the Lens gRPC server
type Client struct {
	pb.IndexerAPIClient
}

// NewClient is used to generate our lens client
func NewClient(opts config.Endpoints) (*Client, error) {
	dialOpts := make([]grpc.DialOption, 0)
	// setup parameters for our conection to
	if opts.Lens.TLS.CertPath != "" {
		creds, err := credentials.NewClientTLSFromFile(opts.Lens.TLS.CertPath, "")
		if err != nil {
			return nil, fmt.Errorf("could not load tls cert: %s", err)
		}
		dialOpts = append(dialOpts,
			grpc.WithTransportCredentials(creds),
			grpc.WithPerRPCCredentials(dialer.NewCredentials(opts.Lens.AuthKey, false)))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}
	var url string
	if opts.Lens.URL == "" {
		url = defaultURL
	}
	gConn, err := grpc.Dial(url, dialOpts...)
	if err != nil {
		return nil, err
	}
	client := &Client{}
	client.IndexerAPIClient = pb.NewIndexerAPIClient(gConn)
	return client, nil
}
