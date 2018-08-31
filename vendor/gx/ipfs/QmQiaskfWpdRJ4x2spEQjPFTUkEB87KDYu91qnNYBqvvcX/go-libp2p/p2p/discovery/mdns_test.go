package discovery

import (
	"context"
	"testing"
	"time"

	bhost "gx/ipfs/QmQiaskfWpdRJ4x2spEQjPFTUkEB87KDYu91qnNYBqvvcX/go-libp2p/p2p/host/basic"

	swarmt "gx/ipfs/QmPWNZRUybw3nwJH3mpkrwB97YEQmXRkzvyh34rpJiih6Q/go-libp2p-swarm/testing"
	host "gx/ipfs/QmRRCrNRs4qxotXx7WJT6SpCvSNEhXvyBcVjXY2K71pcjE/go-libp2p-host"

	pstore "gx/ipfs/QmeKD8YT7887Xu6Z86iZmpYNxrLogJexqxEugSmaf14k64/go-libp2p-peerstore"
)

type DiscoveryNotifee struct {
	h host.Host
}

func (n *DiscoveryNotifee) HandlePeerFound(pi pstore.PeerInfo) {
	n.h.Connect(context.Background(), pi)
}

func TestMdnsDiscovery(t *testing.T) {
	//TODO: re-enable when the new lib will get integrated
	t.Skip("TestMdnsDiscovery fails randomly with current lib")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a := bhost.New(swarmt.GenSwarm(t, ctx))
	b := bhost.New(swarmt.GenSwarm(t, ctx))

	sa, err := NewMdnsService(ctx, a, time.Second, "someTag")
	if err != nil {
		t.Fatal(err)
	}

	sb, err := NewMdnsService(ctx, b, time.Second, "someTag")
	if err != nil {
		t.Fatal(err)
	}

	_ = sb

	n := &DiscoveryNotifee{a}

	sa.RegisterNotifee(n)

	time.Sleep(time.Second * 2)

	err = a.Connect(ctx, pstore.PeerInfo{ID: b.ID()})
	if err != nil {
		t.Fatal(err)
	}
}
