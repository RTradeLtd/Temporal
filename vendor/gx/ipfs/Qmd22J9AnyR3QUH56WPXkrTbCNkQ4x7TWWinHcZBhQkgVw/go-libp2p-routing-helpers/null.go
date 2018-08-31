package routinghelpers

import (
	"context"

	routing "gx/ipfs/QmS4niovD1U6pRjUBXivr1zvvLBqiTKbERjFo994JU7oQS/go-libp2p-routing"
	ropts "gx/ipfs/QmS4niovD1U6pRjUBXivr1zvvLBqiTKbERjFo994JU7oQS/go-libp2p-routing/options"

	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	cid "gx/ipfs/QmZFbDTY9jfSBms2MchvYM9oYRbAF19K7Pby47yDBfpPrb/go-cid"
	pstore "gx/ipfs/QmeKD8YT7887Xu6Z86iZmpYNxrLogJexqxEugSmaf14k64/go-libp2p-peerstore"
)

// Null is a router that doesn't do anything.
type Null struct{}

// PutValue always returns ErrNotSupported
func (nr Null) PutValue(context.Context, string, []byte, ...ropts.Option) error {
	return routing.ErrNotSupported
}

// GetValue always returns ErrNotFound
func (nr Null) GetValue(context.Context, string, ...ropts.Option) ([]byte, error) {
	return nil, routing.ErrNotFound
}

// Provide always returns ErrNotSupported
func (nr Null) Provide(context.Context, *cid.Cid, bool) error {
	return routing.ErrNotSupported
}

// FindProvidersAsync always returns a closed channel
func (nr Null) FindProvidersAsync(context.Context, *cid.Cid, int) <-chan pstore.PeerInfo {
	ch := make(chan pstore.PeerInfo)
	close(ch)
	return ch
}

// FindPeer always returns ErrNotFound
func (nr Null) FindPeer(context.Context, peer.ID) (pstore.PeerInfo, error) {
	return pstore.PeerInfo{}, routing.ErrNotFound
}

// Bootstrap always succeeds instantly
func (nr Null) Bootstrap(context.Context) error {
	return nil
}

var _ routing.IpfsRouting = Null{}
