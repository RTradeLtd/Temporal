package v3

import (
	"context"

	"go.uber.org/zap"

	"github.com/RTradeLtd/Temporal/api/v3/proto/ipfs"
)

// IPFSService implements TemporalIPFSService
type IPFSService struct {
	l *zap.SugaredLogger
}

// CreateNetwork creates a new hosted IPFS network
func (i *IPFSService) CreateNetwork(context.Context, *ipfs.CreateNetworkReq) (*ipfs.NetworkDetails, error) {
	return nil, nil
}

// StartNetwork spins up a previously created IPFS network
func (i *IPFSService) StartNetwork(context.Context, *ipfs.Network) (*ipfs.Empty, error) {
	return nil, nil
}

// StopNetwork shuts down an IPFS network
func (i *IPFSService) StopNetwork(context.Context, *ipfs.Network) (*ipfs.Empty, error) {
	return nil, nil
}

// RemoveNetwork deletes an IPFS network
func (i *IPFSService) RemoveNetwork(context.Context, *ipfs.Network) (*ipfs.Empty, error) {
	return nil, nil
}

// NetworkInfo retrieves information about a network
func (i *IPFSService) NetworkInfo(context.Context, *ipfs.Network) (*ipfs.NetworkDetails, error) {
	return nil, nil
}

// ListNetworks retrieves a list of the authenticated user's networks
func (i *IPFSService) ListNetworks(context.Context, *ipfs.Empty) (*ipfs.NetworkList, error) {
	return nil, nil
}
