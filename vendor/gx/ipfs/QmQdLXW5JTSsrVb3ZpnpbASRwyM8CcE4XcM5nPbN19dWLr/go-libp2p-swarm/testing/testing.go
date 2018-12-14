package testing

import (
	"context"
	"testing"

	inet "gx/ipfs/QmPtFaR7BWHLAjSwLh9kXcyrgTzDpuhcWLkx8ioa9RMYnx/go-libp2p-net"
	tu "gx/ipfs/QmPuhRE325DR8ChNcFtgd6F1eANCHy1oohXZPpYop4xsK6/go-testutil"
	msmux "gx/ipfs/QmRYdszNNq7ykPqavVNKMVyyjX59AcTisHqzussDfhwHkK/go-smux-multistream"
	tcp "gx/ipfs/QmVJrp38KHSdysXZyWq3wK9RtdYW843oa9Psf9MVU6NNx3/go-tcp-transport"
	csms "gx/ipfs/QmX8fjWTb1hynPkHrssWc6YGwphXmfnikLa6VxQzQjMRKJ/go-conn-security-multistream"
	pstore "gx/ipfs/QmZ9zH2FnLcxv1xyzFeUpDUeo55xEhZQHgveZijcxr7TLj/go-libp2p-peerstore"
	pstoremem "gx/ipfs/QmZ9zH2FnLcxv1xyzFeUpDUeo55xEhZQHgveZijcxr7TLj/go-libp2p-peerstore/pstoremem"
	secio "gx/ipfs/QmaTTgNexJg8UVJdJkPzJdNuqGtKDK9BsK31iqK9tn7m8Z/go-libp2p-secio"
	metrics "gx/ipfs/QmbYN6UmTJn5UUQdi5CTsU86TXVBSrTcRk5UmyA36Qx2J6/go-libp2p-metrics"
	tptu "gx/ipfs/Qmc9KUyhx1adPnHX2TBjEWvKej2Gg2kvisAFoQ74UiWYhd/go-libp2p-transport-upgrader"
	yamux "gx/ipfs/Qmdps3CYh5htGQSrPvzg5PHouVexLmtpbuLCqc4vuej8PC/go-smux-yamux"

	swarm "gx/ipfs/QmQdLXW5JTSsrVb3ZpnpbASRwyM8CcE4XcM5nPbN19dWLr/go-libp2p-swarm"
)

type config struct {
	disableReuseport bool
	dialOnly         bool
}

// Option is an option that can be passed when constructing a test swarm.
type Option func(*testing.T, *config)

// OptDisableReuseport disables reuseport in this test swarm.
var OptDisableReuseport Option = func(_ *testing.T, c *config) {
	c.disableReuseport = true
}

// OptDialOnly prevents the test swarm from listening.
var OptDialOnly Option = func(_ *testing.T, c *config) {
	c.dialOnly = true
}

// GenUpgrader creates a new connection upgrader for use with this swarm.
func GenUpgrader(n *swarm.Swarm) *tptu.Upgrader {
	id := n.LocalPeer()
	pk := n.Peerstore().PrivKey(id)
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(secio.ID, &secio.Transport{
		LocalID:    id,
		PrivateKey: pk,
	})

	stMuxer := msmux.NewBlankTransport()
	stMuxer.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)

	return &tptu.Upgrader{
		Secure:  secMuxer,
		Muxer:   stMuxer,
		Filters: n.Filters,
	}

}

// GenSwarm generates a new test swarm.
func GenSwarm(t *testing.T, ctx context.Context, opts ...Option) *swarm.Swarm {
	var cfg config
	for _, o := range opts {
		o(t, &cfg)
	}

	p := tu.RandPeerNetParamsOrFatal(t)

	ps := pstoremem.NewPeerstore()
	ps.AddPubKey(p.ID, p.PubKey)
	ps.AddPrivKey(p.ID, p.PrivKey)
	s := swarm.NewSwarm(ctx, p.ID, ps, metrics.NewBandwidthCounter())

	tcpTransport := tcp.NewTCPTransport(GenUpgrader(s))
	tcpTransport.DisableReuseport = cfg.disableReuseport

	if err := s.AddTransport(tcpTransport); err != nil {
		t.Fatal(err)
	}

	if !cfg.dialOnly {
		if err := s.Listen(p.Addr); err != nil {
			t.Fatal(err)
		}

		s.Peerstore().AddAddrs(p.ID, s.ListenAddresses(), pstore.PermanentAddrTTL)
	}

	return s
}

// DivulgeAddresses adds swarm a's addresses to swarm b's peerstore.
func DivulgeAddresses(a, b inet.Network) {
	id := a.LocalPeer()
	addrs := a.Peerstore().Addrs(id)
	b.Peerstore().AddAddrs(id, addrs, pstore.PermanentAddrTTL)
}
