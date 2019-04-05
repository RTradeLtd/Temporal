package v3

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/RTradeLtd/Temporal/api/v3/proto/ipfs"
)

// IPFSService implements TemporalIPFSService
type IPFSService struct {
	l *zap.SugaredLogger
}

// CreateNetwork creates a new hosted IPFS network
func (i *IPFSService) CreateNetwork(context.Context, *ipfs.CreateNetworkReq) (*ipfs.NetworkDetails, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// StartNetwork spins up a previously created IPFS network
func (i *IPFSService) StartNetwork(context.Context, *ipfs.Network) (*ipfs.Empty, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// StopNetwork shuts down an IPFS network
func (i *IPFSService) StopNetwork(context.Context, *ipfs.Network) (*ipfs.Empty, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// RemoveNetwork deletes an IPFS network
func (i *IPFSService) RemoveNetwork(context.Context, *ipfs.Network) (*ipfs.Empty, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// NetworkInfo retrieves information about a network
func (i *IPFSService) NetworkInfo(context.Context, *ipfs.Network) (*ipfs.NetworkDetails, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}

// ListNetworks retrieves a list of the authenticated user's networks
func (i *IPFSService) ListNetworks(context.Context, *ipfs.Empty) (*ipfs.NetworkList, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "RPC not implemented - coming soon!")
}
