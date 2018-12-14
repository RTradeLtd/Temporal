package routinghelpers

import (
	"context"

	routing "gx/ipfs/QmRASJXJUFygM5qU4YrH7k7jD6S4Hg8nJmgqJ4bYJvLatd/go-libp2p-routing"
	ropts "gx/ipfs/QmRASJXJUFygM5qU4YrH7k7jD6S4Hg8nJmgqJ4bYJvLatd/go-libp2p-routing/options"

	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	pstore "gx/ipfs/QmZ9zH2FnLcxv1xyzFeUpDUeo55xEhZQHgveZijcxr7TLj/go-libp2p-peerstore"
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

// SearchValue always returns ErrNotFound
func (nr Null) SearchValue(ctx context.Context, key string, opts ...ropts.Option) (<-chan []byte, error) {
	return nil, routing.ErrNotFound
}

// Provide always returns ErrNotSupported
func (nr Null) Provide(context.Context, cid.Cid, bool) error {
	return routing.ErrNotSupported
}

// FindProvidersAsync always returns a closed channel
func (nr Null) FindProvidersAsync(context.Context, cid.Cid, int) <-chan pstore.PeerInfo {
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
