package routinghelpers

import (
	"context"

	routing "gx/ipfs/QmS4niovD1U6pRjUBXivr1zvvLBqiTKbERjFo994JU7oQS/go-libp2p-routing"
	ropts "gx/ipfs/QmS4niovD1U6pRjUBXivr1zvvLBqiTKbERjFo994JU7oQS/go-libp2p-routing/options"

	ci "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	peer "gx/ipfs/QmQsErDt8Qgw1XrsXf2BpEzDgGWtB1YLsTAARBup5b6B9W/go-libp2p-peer"
	cid "gx/ipfs/QmZFbDTY9jfSBms2MchvYM9oYRbAF19K7Pby47yDBfpPrb/go-cid"
	pstore "gx/ipfs/QmeKD8YT7887Xu6Z86iZmpYNxrLogJexqxEugSmaf14k64/go-libp2p-peerstore"
	multierror "gx/ipfs/QmfGQp6VVqdPCDyzEM6EGwMY74YPabTSEoQWHUxZuCSWj3/go-multierror"
)

// Tiered is like the Parallel except that GetValue and FindPeer
// are called in series.
type Tiered []routing.IpfsRouting

func (r Tiered) PutValue(ctx context.Context, key string, value []byte, opts ...ropts.Option) error {
	return Parallel(r).PutValue(ctx, key, value, opts...)
}

func (r Tiered) get(ctx context.Context, do func(routing.IpfsRouting) (interface{}, error)) (interface{}, error) {
	var errs []error
	for _, ri := range r {
		val, err := do(ri)
		switch err {
		case nil:
			return val, nil
		case routing.ErrNotFound, routing.ErrNotSupported:
			continue
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		errs = append(errs, err)
	}
	switch len(errs) {
	case 0:
		return nil, routing.ErrNotFound
	case 1:
		return nil, errs[0]
	default:
		return nil, &multierror.Error{Errors: errs}
	}
}

func (r Tiered) GetValue(ctx context.Context, key string, opts ...ropts.Option) ([]byte, error) {
	valInt, err := r.get(ctx, func(ri routing.IpfsRouting) (interface{}, error) {
		return ri.GetValue(ctx, key, opts...)
	})
	val, _ := valInt.([]byte)
	return val, err
}

func (r Tiered) GetPublicKey(ctx context.Context, p peer.ID) (ci.PubKey, error) {
	vInt, err := r.get(ctx, func(ri routing.IpfsRouting) (interface{}, error) {
		return routing.GetPublicKey(ri, ctx, p)
	})
	val, _ := vInt.(ci.PubKey)
	return val, err
}

func (r Tiered) Provide(ctx context.Context, c *cid.Cid, local bool) error {
	return Parallel(r).Provide(ctx, c, local)
}

func (r Tiered) FindProvidersAsync(ctx context.Context, c *cid.Cid, count int) <-chan pstore.PeerInfo {
	return Parallel(r).FindProvidersAsync(ctx, c, count)
}

func (r Tiered) FindPeer(ctx context.Context, p peer.ID) (pstore.PeerInfo, error) {
	valInt, err := r.get(ctx, func(ri routing.IpfsRouting) (interface{}, error) {
		return ri.FindPeer(ctx, p)
	})
	val, _ := valInt.(pstore.PeerInfo)
	return val, err
}

func (r Tiered) Bootstrap(ctx context.Context) error {
	return Parallel(r).Bootstrap(ctx)
}

var _ routing.IpfsRouting = (Tiered)(nil)
