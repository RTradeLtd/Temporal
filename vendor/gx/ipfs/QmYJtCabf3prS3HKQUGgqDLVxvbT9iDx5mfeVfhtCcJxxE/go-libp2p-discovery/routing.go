package discovery

import (
	"context"
	"time"

	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	routing "gx/ipfs/QmRASJXJUFygM5qU4YrH7k7jD6S4Hg8nJmgqJ4bYJvLatd/go-libp2p-routing"
	pstore "gx/ipfs/QmZ9zH2FnLcxv1xyzFeUpDUeo55xEhZQHgveZijcxr7TLj/go-libp2p-peerstore"
	mh "gx/ipfs/QmerPMzPk1mJVowm8KgmoknWa4yCYvvugMPsgWmDNUvDLW/go-multihash"
)

// RoutingDiscovery is an implementation of discovery using ContentRouting
// Namespaces are translated to Cids using the SHA256 hash.
type RoutingDiscovery struct {
	routing.ContentRouting
}

func NewRoutingDiscovery(router routing.ContentRouting) *RoutingDiscovery {
	return &RoutingDiscovery{router}
}

func (d *RoutingDiscovery) Advertise(ctx context.Context, ns string, opts ...Option) (time.Duration, error) {
	cid, err := nsToCid(ns)
	if err != nil {
		return 0, err
	}

	err = d.Provide(ctx, cid, true)
	if err != nil {
		return 0, err
	}

	// this is the dht provide validity
	return 24 * time.Hour, nil
}

func (d *RoutingDiscovery) FindPeers(ctx context.Context, ns string, opts ...Option) (<-chan pstore.PeerInfo, error) {
	var options Options
	err := options.Apply(opts...)
	if err != nil {
		return nil, err
	}

	limit := options.Limit
	if limit == 0 {
		limit = 100 // that's just arbitrary, but FindProvidersAsync needs a count
	}

	cid, err := nsToCid(ns)
	if err != nil {
		return nil, err
	}

	return d.FindProvidersAsync(ctx, cid, limit), nil
}

func nsToCid(ns string) (cid.Cid, error) {
	h, err := mh.Encode([]byte(ns), mh.SHA2_256)
	if err != nil {
		return cid.Undef, err
	}

	return cid.NewCidV1(cid.Raw, h), nil
}
