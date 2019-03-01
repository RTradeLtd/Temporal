package relay_test

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	libp2p "gx/ipfs/QmRxk6AUaGaKCfzS1xSNRojiAPd7h2ih8GuCdjJBF3Y6GK/go-libp2p"
	relay "gx/ipfs/QmRxk6AUaGaKCfzS1xSNRojiAPd7h2ih8GuCdjJBF3Y6GK/go-libp2p/p2p/host/relay"

	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	cid "gx/ipfs/QmTbxNB1NwDesLmKTscr4udL2tVP7MaxvXnD1D9yX7g3PN/go-cid"
	inet "gx/ipfs/QmY3ArotKMKaL7YGfbQfyDrib6RVraLqZYWXZvVgZktBxp/go-libp2p-net"
	peer "gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer"
	host "gx/ipfs/QmYrWiWM4qtrnCeT3R14jY3ZZyirDNJgwK57q4qFYePgbd/go-libp2p-host"
	routing "gx/ipfs/QmYxUdYY9S6yg5tSPVin5GFTvtfsLauVcr7reHDD3dM8xf/go-libp2p-routing"
	circuit "gx/ipfs/QmZBfqr863PYD7BKbmCFSNmzsqYmtr2DKgzubsQaiTQkMc/go-libp2p-circuit"
	autonat "gx/ipfs/QmZjvy5eqc1T9VbFe5KBiDryUJSRNbSbHZD7JMc6CeUXuu/go-libp2p-autonat"
	autonatpb "gx/ipfs/QmZjvy5eqc1T9VbFe5KBiDryUJSRNbSbHZD7JMc6CeUXuu/go-libp2p-autonat/pb"
	pstore "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore"
	manet "gx/ipfs/Qmc85NSvmSG4Frn9Vb2cBc1rMyULH6D3TNVEfCzSKoUpip/go-multiaddr-net"
	ggio "gx/ipfs/QmddjPSGZb3ieihSseFeCfVRpZzcqczPNsD2DvarSwnjJB/gogo-protobuf/io"
)

// test specific parameters
func init() {
	autonat.AutoNATIdentifyDelay = 500 * time.Millisecond
	autonat.AutoNATBootDelay = 1 * time.Second
	relay.BootDelay = 1 * time.Second
	relay.AdvertiseBootDelay = 1 * time.Millisecond
	manet.Private4 = []*net.IPNet{}
}

// mock routing
type mockRoutingTable struct {
	mx        sync.Mutex
	providers map[string]map[peer.ID]pstore.PeerInfo
}

type mockRouting struct {
	h   host.Host
	tab *mockRoutingTable
}

func newMockRoutingTable() *mockRoutingTable {
	return &mockRoutingTable{providers: make(map[string]map[peer.ID]pstore.PeerInfo)}
}

func newMockRouting(h host.Host, tab *mockRoutingTable) *mockRouting {
	return &mockRouting{h: h, tab: tab}
}

func (m *mockRouting) FindPeer(ctx context.Context, p peer.ID) (pstore.PeerInfo, error) {
	return pstore.PeerInfo{}, routing.ErrNotFound
}

func (m *mockRouting) Provide(ctx context.Context, cid cid.Cid, bcast bool) error {
	m.tab.mx.Lock()
	defer m.tab.mx.Unlock()

	pmap, ok := m.tab.providers[cid.String()]
	if !ok {
		pmap = make(map[peer.ID]pstore.PeerInfo)
		m.tab.providers[cid.String()] = pmap
	}

	pmap[m.h.ID()] = pstore.PeerInfo{ID: m.h.ID(), Addrs: m.h.Addrs()}

	return nil
}

func (m *mockRouting) FindProvidersAsync(ctx context.Context, cid cid.Cid, limit int) <-chan pstore.PeerInfo {
	ch := make(chan pstore.PeerInfo)
	go func() {
		defer close(ch)
		m.tab.mx.Lock()
		defer m.tab.mx.Unlock()

		pmap, ok := m.tab.providers[cid.String()]
		if !ok {
			return
		}

		for _, pi := range pmap {
			select {
			case ch <- pi:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch
}

// mock autonat
func makeAutoNATServicePrivate(ctx context.Context, t *testing.T) host.Host {
	h, err := libp2p.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	h.SetStreamHandler(autonat.AutoNATProto, sayAutoNATPrivate)
	return h
}

func sayAutoNATPrivate(s inet.Stream) {
	defer s.Close()
	w := ggio.NewDelimitedWriter(s)
	res := autonatpb.Message{
		Type:         autonatpb.Message_DIAL_RESPONSE.Enum(),
		DialResponse: newDialResponseError(autonatpb.Message_E_DIAL_ERROR, "no dialable addresses"),
	}
	w.WriteMsg(&res)
}

func newDialResponseError(status autonatpb.Message_ResponseStatus, text string) *autonatpb.Message_DialResponse {
	dr := new(autonatpb.Message_DialResponse)
	dr.Status = status.Enum()
	dr.StatusText = &text
	return dr
}

// connector
func connect(t *testing.T, a, b host.Host) {
	pinfo := pstore.PeerInfo{ID: a.ID(), Addrs: a.Addrs()}
	err := b.Connect(context.Background(), pinfo)
	if err != nil {
		t.Fatal(err)
	}
}

// and the actual test!
func TestAutoRelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mtab := newMockRoutingTable()
	makeRouting := func(h host.Host) (routing.PeerRouting, error) {
		mr := newMockRouting(h, mtab)
		return mr, nil
	}

	h1 := makeAutoNATServicePrivate(ctx, t)
	_, err := libp2p.New(ctx, libp2p.EnableRelay(circuit.OptHop), libp2p.EnableAutoRelay(), libp2p.Routing(makeRouting))
	if err != nil {
		t.Fatal(err)
	}
	h3, err := libp2p.New(ctx, libp2p.EnableRelay(), libp2p.EnableAutoRelay(), libp2p.Routing(makeRouting))
	if err != nil {
		t.Fatal(err)
	}
	h4, err := libp2p.New(ctx, libp2p.EnableRelay())

	// verify that we don't advertise relay addrs initially
	for _, addr := range h3.Addrs() {
		_, err := addr.ValueForProtocol(circuit.P_CIRCUIT)
		if err == nil {
			t.Fatal("relay addr advertised before auto detection")
		}
	}

	// connect to AutoNAT and let detection/discovery work its magic
	connect(t, h1, h3)
	time.Sleep(3 * time.Second)

	// verify that we now advertise relay addrs (but not unspecific relay addrs)
	unspecificRelay, err := ma.NewMultiaddr("/p2p-circuit")
	if err != nil {
		t.Fatal(err)
	}

	haveRelay := false
	for _, addr := range h3.Addrs() {
		if addr.Equal(unspecificRelay) {
			t.Fatal("unspecific relay addr advertised")
		}

		_, err := addr.ValueForProtocol(circuit.P_CIRCUIT)
		if err == nil {
			haveRelay = true
		}
	}

	if !haveRelay {
		t.Fatal("No relay addrs advertised")
	}

	// verify that we can connect through the relay
	var raddrs []ma.Multiaddr
	for _, addr := range h3.Addrs() {
		_, err := addr.ValueForProtocol(circuit.P_CIRCUIT)
		if err == nil {
			raddrs = append(raddrs, addr)
		}
	}

	err = h4.Connect(ctx, pstore.PeerInfo{ID: h3.ID(), Addrs: raddrs})
	if err != nil {
		t.Fatal(err)
	}

	// verify that we have pushed relay addrs to connected peers
	haveRelay = false
	for _, addr := range h1.Peerstore().Addrs(h3.ID()) {
		if addr.Equal(unspecificRelay) {
			t.Fatal("unspecific relay addr advertised")
		}

		_, err := addr.ValueForProtocol(circuit.P_CIRCUIT)
		if err == nil {
			haveRelay = true
		}
	}

	if !haveRelay {
		t.Fatal("No relay addrs pushed")
	}
}
