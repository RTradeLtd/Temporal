package temporal

import (
	"google.golang.org/grpc"

	"github.com/RTradeLtd/sdk/go/temporal/auth"
	"github.com/RTradeLtd/sdk/go/temporal/ipfs"
	"github.com/RTradeLtd/sdk/go/temporal/store"
)

// MetaKey denotes context metadata keys used by the Temporal gRPC API
type MetaKey string

const (
	// MetaKeyAuthorization is the key for authorization tokens
	MetaKeyAuthorization MetaKey = "authorization"

	// BlobThreshold is the threshold in bytes before blobs are required to be broken up
	BlobThreshold = 5e+6

	apiAddress = "api.temporal.cloud"
)

// NewAuthClient instantiates a new client for Temporal's authentication APIs
func NewAuthClient(conn *grpc.ClientConn) auth.TemporalAuthClient {
	return auth.NewTemporalAuthClient(conn)
}

// NewStoreClient instantiates a new client for Temporal's storage APIs
func NewStoreClient(conn *grpc.ClientConn) store.TemporalStoreClient {
	return store.NewTemporalStoreClient(conn)
}

// NewIPFSClient instantiates a new client for Temporal's IPFS APIs
func NewIPFSClient(conn *grpc.ClientConn) ipfs.TemporalIPFSClient {
	return ipfs.NewTemporalIPFSClient(conn)
}
