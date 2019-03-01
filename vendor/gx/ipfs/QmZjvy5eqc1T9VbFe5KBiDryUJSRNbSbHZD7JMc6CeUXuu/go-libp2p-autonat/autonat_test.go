package autonat

import (
	"context"
	"testing"
	"time"

	pb "gx/ipfs/QmZjvy5eqc1T9VbFe5KBiDryUJSRNbSbHZD7JMc6CeUXuu/go-libp2p-autonat/pb"

	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	inet "gx/ipfs/QmY3ArotKMKaL7YGfbQfyDrib6RVraLqZYWXZvVgZktBxp/go-libp2p-net"
	host "gx/ipfs/QmYrWiWM4qtrnCeT3R14jY3ZZyirDNJgwK57q4qFYePgbd/go-libp2p-host"
	swarmt "gx/ipfs/Qma3Xp3FXFSP4prirEiRYHJ2tgGE8EAx9i6JLziPLpAQjq/go-libp2p-swarm/testing"
	pstore "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore"
	bhost "gx/ipfs/QmcBTHN7uAMBdkzRoQ3n9cE7tGu8Ubd9zmahjskjTRw4Uf/go-libp2p-blankhost"
	ggio "gx/ipfs/QmddjPSGZb3ieihSseFeCfVRpZzcqczPNsD2DvarSwnjJB/gogo-protobuf/io"
)

func init() {
	AutoNATBootDelay = 1 * time.Second
	AutoNATRefreshInterval = 1 * time.Second
	AutoNATRetryInterval = 1 * time.Second
	AutoNATIdentifyDelay = 100 * time.Millisecond
}

// these are mock service implementations for testing
func makeAutoNATServicePrivate(ctx context.Context, t *testing.T) host.Host {
	h := bhost.NewBlankHost(swarmt.GenSwarm(t, ctx))
	h.SetStreamHandler(AutoNATProto, sayAutoNATPrivate)
	return h
}

func makeAutoNATServicePublic(ctx context.Context, t *testing.T) host.Host {
	h := bhost.NewBlankHost(swarmt.GenSwarm(t, ctx))
	h.SetStreamHandler(AutoNATProto, sayAutoNATPublic)
	return h
}

func sayAutoNATPrivate(s inet.Stream) {
	defer s.Close()
	w := ggio.NewDelimitedWriter(s)
	res := pb.Message{
		Type:         pb.Message_DIAL_RESPONSE.Enum(),
		DialResponse: newDialResponseError(pb.Message_E_DIAL_ERROR, "no dialable addresses"),
	}
	w.WriteMsg(&res)
}

func sayAutoNATPublic(s inet.Stream) {
	defer s.Close()
	w := ggio.NewDelimitedWriter(s)
	res := pb.Message{
		Type:         pb.Message_DIAL_RESPONSE.Enum(),
		DialResponse: newDialResponseOK(s.Conn().RemoteMultiaddr()),
	}
	w.WriteMsg(&res)
}

func newDialResponseOK(addr ma.Multiaddr) *pb.Message_DialResponse {
	dr := new(pb.Message_DialResponse)
	dr.Status = pb.Message_OK.Enum()
	dr.Addr = addr.Bytes()
	return dr
}

func newDialResponseError(status pb.Message_ResponseStatus, text string) *pb.Message_DialResponse {
	dr := new(pb.Message_DialResponse)
	dr.Status = status.Enum()
	dr.StatusText = &text
	return dr
}

func makeAutoNAT(ctx context.Context, t *testing.T, ash host.Host) (host.Host, AutoNAT) {
	h := bhost.NewBlankHost(swarmt.GenSwarm(t, ctx))
	a := NewAutoNAT(ctx, h, nil)
	a.(*AmbientAutoNAT).peers[ash.ID()] = ash.Addrs()

	return h, a
}

func connect(t *testing.T, a, b host.Host) {
	pinfo := pstore.PeerInfo{ID: a.ID(), Addrs: a.Addrs()}
	err := b.Connect(context.Background(), pinfo)
	if err != nil {
		t.Fatal(err)
	}
}

// tests
func TestAutoNATPrivate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hs := makeAutoNATServicePrivate(ctx, t)
	hc, an := makeAutoNAT(ctx, t, hs)

	status := an.Status()
	if status != NATStatusUnknown {
		t.Fatalf("unexpected NAT status: %d", status)
	}

	connect(t, hs, hc)
	time.Sleep(2 * time.Second)

	status = an.Status()
	if status != NATStatusPrivate {
		t.Fatalf("unexpected NAT status: %d", status)
	}
}

func TestAutoNATPublic(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hs := makeAutoNATServicePublic(ctx, t)
	hc, an := makeAutoNAT(ctx, t, hs)

	status := an.Status()
	if status != NATStatusUnknown {
		t.Fatalf("unexpected NAT status: %d", status)
	}

	connect(t, hs, hc)
	time.Sleep(2 * time.Second)

	status = an.Status()
	if status != NATStatusPublic {
		t.Fatalf("unexpected NAT status: %d", status)
	}
}
