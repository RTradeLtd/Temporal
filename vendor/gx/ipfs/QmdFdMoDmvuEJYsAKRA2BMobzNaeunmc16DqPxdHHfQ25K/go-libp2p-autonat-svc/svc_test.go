package autonat

import (
	"context"
	"net"
	"testing"
	"time"

	manet "gx/ipfs/QmQVUtnrNGtCRkCMpXgpApfzQjc8FDaDVxHqWH8cnZQeh5/go-multiaddr-net"
	autonat "gx/ipfs/QmQXeTmQcPFnf3ZAvik2qgKxWNoZ27aKGcW8hUBwBrTxT1/go-libp2p-autonat"
	libp2p "gx/ipfs/QmRBaUEQEeFWywfrZJ64QgsmvcqgLSK3VbvGMR2NM2Edpf/go-libp2p"
	pstore "gx/ipfs/QmZ9zH2FnLcxv1xyzFeUpDUeo55xEhZQHgveZijcxr7TLj/go-libp2p-peerstore"
	host "gx/ipfs/QmfD51tKgJiTMnW9JEiDiPwsCY4mqUoxkhKhBfyW12spTC/go-libp2p-host"
)

func makeAutoNATService(ctx context.Context, t *testing.T) (host.Host, *AutoNATService) {
	h, err := libp2p.New(ctx, libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		t.Fatal(err)
	}

	as, err := NewAutoNATService(ctx, h)
	if err != nil {
		t.Fatal(err)
	}

	return h, as
}

func makeAutoNATClient(ctx context.Context, t *testing.T) (host.Host, autonat.AutoNATClient) {
	h, err := libp2p.New(ctx, libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"))
	if err != nil {
		t.Fatal(err)
	}

	cli := autonat.NewAutoNATClient(h, nil)
	return h, cli
}

func connect(t *testing.T, a, b host.Host) {
	pinfo := pstore.PeerInfo{ID: a.ID(), Addrs: a.Addrs()}
	err := b.Connect(context.Background(), pinfo)
	if err != nil {
		t.Fatal(err)
	}
}

// Note: these tests assume that the host has only private inet addresses!
func TestAutoNATServiceDialError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	save := AutoNATServiceDialTimeout
	AutoNATServiceDialTimeout = 1 * time.Second

	hs, _ := makeAutoNATService(ctx, t)
	hc, ac := makeAutoNATClient(ctx, t)
	connect(t, hs, hc)

	_, err := ac.DialBack(ctx, hs.ID())
	if err == nil {
		t.Fatal("Dial back succeeded unexpectedly!")
	}

	if !autonat.IsDialError(err) {
		t.Fatal(err)
	}

	AutoNATServiceDialTimeout = save
}

func TestAutoNATServiceDialSuccess(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	save := manet.Private4
	manet.Private4 = []*net.IPNet{}

	hs, _ := makeAutoNATService(ctx, t)
	hc, ac := makeAutoNATClient(ctx, t)
	connect(t, hs, hc)

	_, err := ac.DialBack(ctx, hs.ID())
	if err != nil {
		t.Fatalf("Dial back failed: %s", err.Error())
	}

	manet.Private4 = save
}

func TestAutoNATServiceDialRateLimiter(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	save1 := AutoNATServiceDialTimeout
	AutoNATServiceDialTimeout = 1 * time.Second
	save2 := AutoNATServiceResetInterval
	AutoNATServiceResetInterval = 1 * time.Second
	save3 := AutoNATServiceThrottle
	AutoNATServiceThrottle = 1
	save4 := manet.Private4
	manet.Private4 = []*net.IPNet{}

	hs, _ := makeAutoNATService(ctx, t)
	hc, ac := makeAutoNATClient(ctx, t)
	connect(t, hs, hc)

	_, err := ac.DialBack(ctx, hs.ID())
	if err != nil {
		t.Fatal(err)
	}

	_, err = ac.DialBack(ctx, hs.ID())
	if err == nil {
		t.Fatal("Dial back succeeded unexpectedly!")
	}

	if !autonat.IsDialRefused(err) {
		t.Fatal(err)
	}

	time.Sleep(2 * time.Second)

	_, err = ac.DialBack(ctx, hs.ID())
	if err != nil {
		t.Fatal(err)
	}

	AutoNATServiceDialTimeout = save1
	AutoNATServiceResetInterval = save2
	AutoNATServiceThrottle = save3
	manet.Private4 = save4
}
