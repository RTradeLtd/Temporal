package swarm_test

import (
	"context"
	"testing"

	swarmt "gx/ipfs/QmQdLXW5JTSsrVb3ZpnpbASRwyM8CcE4XcM5nPbN19dWLr/go-libp2p-swarm/testing"

	ma "gx/ipfs/QmRKLtwMw131aK7ugC3G7ybpumMz78YrJe5dzneyindvG1/go-multiaddr"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	transport "gx/ipfs/Qmb3qartY8DSgRaBA3Go4EEjY1ZbXhCcvmc4orsBKMjgRg/go-libp2p-transport"
)

type dummyTransport struct {
	protocols []int
	proxy     bool
}

func (dt *dummyTransport) Dial(ctx context.Context, raddr ma.Multiaddr, p peer.ID) (transport.Conn, error) {
	panic("unimplemented")
}

func (dt *dummyTransport) CanDial(addr ma.Multiaddr) bool {
	panic("unimplemented")
}

func (dt *dummyTransport) Listen(laddr ma.Multiaddr) (transport.Listener, error) {
	panic("unimplemented")
}

func (dt *dummyTransport) Proxy() bool {
	return dt.proxy
}

func (dt *dummyTransport) Protocols() []int {
	return dt.protocols
}

func TestUselessTransport(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	swarm := swarmt.GenSwarm(t, ctx)
	err := swarm.AddTransport(new(dummyTransport))
	if err == nil {
		t.Fatal("adding a transport that supports no protocols should have failed")
	}
}
