package libp2p

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	crypto "gx/ipfs/QmPvyPwuCgJ7pDmrKDxRtsScJgBaM5h4EpRL2qQJsmXf4n/go-libp2p-crypto"
	host "gx/ipfs/QmRRCrNRs4qxotXx7WJT6SpCvSNEhXvyBcVjXY2K71pcjE/go-libp2p-host"
	"gx/ipfs/QmcK89iqkFV8TqpRUgx1481YZbhjPFnBjqkpBQJfJqmSfm/go-tcp-transport"
)

func TestNewHost(t *testing.T) {
	h, err := makeRandomHost(t, 9000)
	if err != nil {
		t.Fatal(err)
	}
	h.Close()
}

func TestBadTransportConstructor(t *testing.T) {
	ctx := context.Background()
	h, err := New(ctx, Transport(func() {}))
	if err == nil {
		h.Close()
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "libp2p_test.go") {
		t.Error("expected error to contain debugging info")
	}
}

func TestInsecure(t *testing.T) {
	ctx := context.Background()
	h, err := New(ctx, NoSecurity)
	if err != nil {
		t.Fatal(err)
	}
	h.Close()
}

func TestDefaultListenAddrs(t *testing.T) {
	ctx := context.Background()

	re := regexp.MustCompile("/(ip)[4|6]/((0.0.0.0)|(::))/tcp/")

	// Test 1: Setting the correct listen addresses if userDefined.Transport == nil && userDefined.ListenAddrs == nil
	h, err := New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, addr := range h.Network().ListenAddresses() {
		if re.FindStringSubmatchIndex(addr.String()) == nil {
			t.Error("expected ip4 or ip6 interface")
		}
	}

	h.Close()

	// Test 2: Listen addr should not set if user defined transport is passed.
	h, err = New(
		ctx,
		Transport(tcp.NewTCPTransport),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(h.Network().ListenAddresses()) != 0 {
		t.Error("expected zero listen addrs as none is set with user defined transport")
	}
	h.Close()
}

func makeRandomHost(t *testing.T, port int) (host.Host, error) {
	ctx := context.Background()
	priv, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		t.Fatal(err)
	}

	opts := []Option{
		ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)),
		Identity(priv),
		DefaultTransports,
		DefaultMuxers,
		DefaultSecurity,
		NATPortMap(),
	}

	return New(ctx, opts...)
}
