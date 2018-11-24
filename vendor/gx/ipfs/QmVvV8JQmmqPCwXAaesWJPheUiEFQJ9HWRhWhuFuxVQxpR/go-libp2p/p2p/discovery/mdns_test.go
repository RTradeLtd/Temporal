package discovery

import (
	"context"
	"testing"
	"time"

	bhost "gx/ipfs/QmVvV8JQmmqPCwXAaesWJPheUiEFQJ9HWRhWhuFuxVQxpR/go-libp2p/p2p/host/basic"

	swarmt "gx/ipfs/QmQrYHkcGprZBUFnRigeiZFkaFDBHtmRhDdPpSiiUTRNwv/go-libp2p-swarm/testing"
	host "gx/ipfs/QmahxMNoNuSsgQefo9rkpcfRFmQrMN6Q99aztKXf63K7YJ/go-libp2p-host"

	pstore "gx/ipfs/QmQAGG1zxfePqj2t7bLxyN8AFccZ889DDR9Gn8kVLDrGZo/go-libp2p-peerstore"
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
