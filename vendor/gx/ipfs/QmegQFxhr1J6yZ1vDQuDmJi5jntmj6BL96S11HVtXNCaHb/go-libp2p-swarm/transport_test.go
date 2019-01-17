package swarm_test

import (
	"context"
	"testing"

	swarmt "gx/ipfs/QmegQFxhr1J6yZ1vDQuDmJi5jntmj6BL96S11HVtXNCaHb/go-libp2p-swarm/testing"

	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	transport "gx/ipfs/QmS4UBXoQ5QgTJA5pc62egqa5KrQRhsDHhaFHEoGUASsxp/go-libp2p-transport"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
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
