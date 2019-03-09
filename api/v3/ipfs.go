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

// Keys retrieves an account's IPFS keys
func (i *IPFSService) Keys(ctx context.Context, req *ipfs.Empty) (*ipfs.KeysResp, error) {
	return nil, nil
}

// NewKey generates a new IPFS key associated with an authenticated request.
func (i *IPFSService) NewKey(ctx context.Context, req *ipfs.Key) (*ipfs.Empty, error) {
	return nil, nil
}
