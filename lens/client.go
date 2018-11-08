package lens

import (
	pb "github.com/RTradeLtd/grpc/lens"
	"google.golang.org/grpc"
)

// Client is a lens client used to make requests to the Lens gRPC server
type Client struct {
	pb.IndexerAPIClient
}

// NewClient is used to generate our lens client
func NewClient(url string, insecure bool) (*Client, error) {
	var (
		client *Client
		gconn  *grpc.ClientConn
		err    error
	)
	if insecure {
		gconn, err = grpc.Dial(url, grpc.WithInsecure())
	}
	if err != nil {
		return nil, err
	}
	client.IndexerAPIClient = pb.NewIndexerAPIClient(gconn)
	return client, nil
}
