package testing

import (
	"context"
	"testing"

	tu "gx/ipfs/QmRNhSdqzMcuRxX9A1egBeQ3BhDTguDV5HPwi8wRykkPU8/go-testutil"
	secio "gx/ipfs/QmWKKkNLFRcznF5vDqt2eeRsnQqQhwbjVf8zJ9KC2RXrzN/go-libp2p-secio"
	msmux "gx/ipfs/QmWzjXAyBTygw6CeCTUnhJzhFucfxY5FJivSoiGuiSbPjS/go-smux-multistream"
	inet "gx/ipfs/QmX5J1q63BrrDTbpcHifrFbxH3cMZsvaNajy6u3zCpzBXs/go-libp2p-net"
	csms "gx/ipfs/Qma6UXLMHjdVFExQZLYqdb5KAesbnoXuthQzovrwRZ64fG/go-conn-security-multistream"
	tcp "gx/ipfs/QmcK89iqkFV8TqpRUgx1481YZbhjPFnBjqkpBQJfJqmSfm/go-tcp-transport"
	yamux "gx/ipfs/QmcsgrV3nCAKjiHKZhKVXWc4oY3WBECJCqahXEMpHeMrev/go-smux-yamux"
	metrics "gx/ipfs/QmdhwKw53CTV8EJSAsR1bpmMT5kXiWBgeAyv1EXeeDiXqR/go-libp2p-metrics"
	pstore "gx/ipfs/QmeKD8YT7887Xu6Z86iZmpYNxrLogJexqxEugSmaf14k64/go-libp2p-peerstore"
	tptu "gx/ipfs/QmfNvpHX396fhMeauERV6eFnSJg78rUjhjpFf1JvbjxaYM/go-libp2p-transport-upgrader"

	swarm "gx/ipfs/QmPWNZRUybw3nwJH3mpkrwB97YEQmXRkzvyh34rpJiih6Q/go-libp2p-swarm"
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

	ps := pstore.NewPeerstore()
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
