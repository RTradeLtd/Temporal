package relay

import (
	"context"

	basic "gx/ipfs/QmVvV8JQmmqPCwXAaesWJPheUiEFQJ9HWRhWhuFuxVQxpR/go-libp2p/p2p/host/basic"

	ma "gx/ipfs/QmRKLtwMw131aK7ugC3G7ybpumMz78YrJe5dzneyindvG1/go-multiaddr"
	host "gx/ipfs/QmahxMNoNuSsgQefo9rkpcfRFmQrMN6Q99aztKXf63K7YJ/go-libp2p-host"
	discovery "gx/ipfs/QmejQHZsodEXxdeFQazcWsi4Dkmp4mX7QEZrWXHoVR5EtK/go-libp2p-discovery"
)

// RelayHost is a Host that provides Relay services.
type RelayHost struct {
	*basic.BasicHost
	advertise discovery.Advertiser
	addrsF    basic.AddrsFactory
}

// New constructs a new RelayHost
func NewRelayHost(ctx context.Context, bhost *basic.BasicHost, advertise discovery.Advertiser) *RelayHost {
	h := &RelayHost{
		BasicHost: bhost,
		addrsF:    bhost.AddrsFactory,
		advertise: advertise,
	}
	bhost.AddrsFactory = h.hostAddrs
	discovery.Advertise(ctx, advertise, RelayRendezvous)
	return h
}

func (h *RelayHost) hostAddrs(addrs []ma.Multiaddr) []ma.Multiaddr {
	return filterUnspecificRelay(h.addrsF(addrs))
}

var _ host.Host = (*RelayHost)(nil)
