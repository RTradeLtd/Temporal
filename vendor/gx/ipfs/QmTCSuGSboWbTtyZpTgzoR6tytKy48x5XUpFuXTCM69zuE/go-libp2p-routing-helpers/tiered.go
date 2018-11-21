package routinghelpers

import (
	"context"

	routing "github.com/libp2p/go-libp2p-routing"
	ropts "github.com/libp2p/go-libp2p-routing/options"

	multierror "github.com/hashicorp/go-multierror"
	cid "github.com/ipfs/go-cid"
	ci "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	record "github.com/libp2p/go-libp2p-record"
)

// Tiered is like the Parallel except that GetValue and FindPeer
// are called in series.
type Tiered struct {
	Routers   []routing.IpfsRouting
	Validator record.Validator
}

func (r Tiered) PutValue(ctx context.Context, key string, value []byte, opts ...ropts.Option) error {
	return Parallel{Routers: r.Routers}.PutValue(ctx, key, value, opts...)
}

func (r Tiered) get(ctx context.Context, do func(routing.IpfsRouting) (interface{}, error)) (interface{}, error) {
	var errs []error
	for _, ri := range r.Routers {
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

func (r Tiered) SearchValue(ctx context.Context, key string, opts ...ropts.Option) (<-chan []byte, error) {
	return Parallel{Routers: r.Routers, Validator: r.Validator}.SearchValue(ctx, key, opts...)
}

func (r Tiered) GetPublicKey(ctx context.Context, p peer.ID) (ci.PubKey, error) {
	vInt, err := r.get(ctx, func(ri routing.IpfsRouting) (interface{}, error) {
		return routing.GetPublicKey(ri, ctx, p)
	})
	val, _ := vInt.(ci.PubKey)
	return val, err
}

func (r Tiered) Provide(ctx context.Context, c cid.Cid, local bool) error {
	return Parallel{Routers: r.Routers}.Provide(ctx, c, local)
}

func (r Tiered) FindProvidersAsync(ctx context.Context, c cid.Cid, count int) <-chan pstore.PeerInfo {
	return Parallel{Routers: r.Routers}.FindProvidersAsync(ctx, c, count)
}

func (r Tiered) FindPeer(ctx context.Context, p peer.ID) (pstore.PeerInfo, error) {
	valInt, err := r.get(ctx, func(ri routing.IpfsRouting) (interface{}, error) {
		return ri.FindPeer(ctx, p)
	})
	val, _ := valInt.(pstore.PeerInfo)
	return val, err
}

func (r Tiered) Bootstrap(ctx context.Context) error {
	return Parallel{Routers: r.Routers}.Bootstrap(ctx)
}

var _ routing.IpfsRouting = Tiered{}
