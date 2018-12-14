package ping_test

import (
	"context"
	"testing"
	"time"

	swarmt "gx/ipfs/QmQdLXW5JTSsrVb3ZpnpbASRwyM8CcE4XcM5nPbN19dWLr/go-libp2p-swarm/testing"
	bhost "gx/ipfs/QmRBaUEQEeFWywfrZJ64QgsmvcqgLSK3VbvGMR2NM2Edpf/go-libp2p/p2p/host/basic"
	ping "gx/ipfs/QmRBaUEQEeFWywfrZJ64QgsmvcqgLSK3VbvGMR2NM2Edpf/go-libp2p/p2p/protocol/ping"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	pstore "gx/ipfs/QmZ9zH2FnLcxv1xyzFeUpDUeo55xEhZQHgveZijcxr7TLj/go-libp2p-peerstore"
)

func TestPing(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h1 := bhost.New(swarmt.GenSwarm(t, ctx))
	h2 := bhost.New(swarmt.GenSwarm(t, ctx))

	err := h1.Connect(ctx, pstore.PeerInfo{
		ID:    h2.ID(),
		Addrs: h2.Addrs(),
	})

	if err != nil {
		t.Fatal(err)
	}

	ps1 := ping.NewPingService(h1)
	ps2 := ping.NewPingService(h2)

	testPing(t, ps1, h2.ID())
	testPing(t, ps2, h1.ID())
}

func testPing(t *testing.T, ps *ping.PingService, p peer.ID) {
	pctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ts, err := ps.Ping(pctx, p)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		select {
		case took := <-ts:
			t.Log("ping took: ", took)
		case <-time.After(time.Second * 4):
			t.Fatal("failed to receive ping")
		}
	}

}
