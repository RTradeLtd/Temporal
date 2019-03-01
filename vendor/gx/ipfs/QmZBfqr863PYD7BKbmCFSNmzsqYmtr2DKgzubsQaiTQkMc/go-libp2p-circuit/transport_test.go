package relay_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	. "gx/ipfs/QmZBfqr863PYD7BKbmCFSNmzsqYmtr2DKgzubsQaiTQkMc/go-libp2p-circuit"

	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	inet "gx/ipfs/QmY3ArotKMKaL7YGfbQfyDrib6RVraLqZYWXZvVgZktBxp/go-libp2p-net"
	host "gx/ipfs/QmYrWiWM4qtrnCeT3R14jY3ZZyirDNJgwK57q4qFYePgbd/go-libp2p-host"
	swarm "gx/ipfs/Qma3Xp3FXFSP4prirEiRYHJ2tgGE8EAx9i6JLziPLpAQjq/go-libp2p-swarm"
	swarmt "gx/ipfs/Qma3Xp3FXFSP4prirEiRYHJ2tgGE8EAx9i6JLziPLpAQjq/go-libp2p-swarm/testing"
	pstore "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore"
)

const TestProto = "test/relay-transport"

var msg = []byte("relay works!")

func testSetupRelay(t *testing.T, ctx context.Context) []host.Host {
	hosts := getNetHosts(t, ctx, 3)

	err := AddRelayTransport(ctx, hosts[0], swarmt.GenUpgrader(hosts[0].Network().(*swarm.Swarm)), OptDiscovery)
	if err != nil {
		t.Fatal(err)
	}

	err = AddRelayTransport(ctx, hosts[1], swarmt.GenUpgrader(hosts[1].Network().(*swarm.Swarm)), OptHop)
	if err != nil {
		t.Fatal(err)
	}

	err = AddRelayTransport(ctx, hosts[2], swarmt.GenUpgrader(hosts[2].Network().(*swarm.Swarm)))
	if err != nil {
		t.Fatal(err)
	}

	connect(t, hosts[0], hosts[1])
	connect(t, hosts[1], hosts[2])

	time.Sleep(100 * time.Millisecond)

	handler := func(s inet.Stream) {
		_, err := s.Write(msg)
		if err != nil {
			t.Error(err)
		}
		s.Close()
	}

	hosts[2].SetStreamHandler(TestProto, handler)

	return hosts
}

func TestFullAddressTransportDial(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts := testSetupRelay(t, ctx)

	addr, err := ma.NewMultiaddr(fmt.Sprintf("%s/ipfs/%s/p2p-circuit/ipfs/%s", hosts[1].Addrs()[0].String(), hosts[1].ID().Pretty(), hosts[2].ID().Pretty()))
	if err != nil {
		t.Fatal(err)
	}

	rctx, rcancel := context.WithTimeout(ctx, time.Second)
	defer rcancel()

	hosts[0].Peerstore().AddAddrs(hosts[2].ID(), []ma.Multiaddr{addr}, pstore.TempAddrTTL)

	s, err := hosts[0].NewStream(rctx, hosts[2].ID(), TestProto)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(s)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, msg) {
		t.Fatal("message was incorrect:", string(data))
	}
}

func TestSpecificRelayTransportDial(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts := testSetupRelay(t, ctx)

	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s/p2p-circuit/ipfs/%s", hosts[1].ID().Pretty(), hosts[2].ID().Pretty()))
	if err != nil {
		t.Fatal(err)
	}

	rctx, rcancel := context.WithTimeout(ctx, time.Second)
	defer rcancel()

	hosts[0].Peerstore().AddAddrs(hosts[2].ID(), []ma.Multiaddr{addr}, pstore.TempAddrTTL)

	s, err := hosts[0].NewStream(rctx, hosts[2].ID(), TestProto)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(s)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, msg) {
		t.Fatal("message was incorrect:", string(data))
	}
}

func TestUnspecificRelayTransportDial(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hosts := testSetupRelay(t, ctx)

	addr, err := ma.NewMultiaddr(fmt.Sprintf("/p2p-circuit/ipfs/%s", hosts[2].ID().Pretty()))
	if err != nil {
		t.Fatal(err)
	}

	rctx, rcancel := context.WithTimeout(ctx, time.Second)
	defer rcancel()

	hosts[0].Peerstore().AddAddrs(hosts[2].ID(), []ma.Multiaddr{addr}, pstore.TempAddrTTL)

	s, err := hosts[0].NewStream(rctx, hosts[2].ID(), TestProto)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadAll(s)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, msg) {
		t.Fatal("message was incorrect:", string(data))
	}
}
