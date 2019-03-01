package swarm_test

import (
	"context"
	"testing"

	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	testutil "gx/ipfs/QmWapVoHjtKhn4MhvKNoPTkJKADFGACfXPFnt7combwp5W/go-testutil"
	pstore "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore"
)

func TestDialBadAddrs(t *testing.T) {

	m := func(s string) ma.Multiaddr {
		maddr, err := ma.NewMultiaddr(s)
		if err != nil {
			t.Fatal(err)
		}
		return maddr
	}

	ctx := context.Background()
	s := makeSwarms(ctx, t, 1)[0]

	test := func(a ma.Multiaddr) {
		p := testutil.RandPeerIDFatal(t)
		s.Peerstore().AddAddr(p, a, pstore.PermanentAddrTTL)
		if _, err := s.DialPeer(ctx, p); err == nil {
			t.Errorf("swarm should not dial: %s", p)
		}
	}

	test(m("/ip6/fe80::1"))                // link local
	test(m("/ip6/fe80::100"))              // link local
	test(m("/ip4/127.0.0.1/udp/1234/utp")) // utp
}
