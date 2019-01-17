package swarm_test

import (
	"context"
	"testing"

	ma "gx/ipfs/QmNTCey11oxhb1AxDnQBRHtdhap6Ctud872NjAYPYYXPuc/go-multiaddr"
	testutil "gx/ipfs/QmNvHv84aH2qZafDuSdKJCQ1cvPZ1kmQmyD4YtzjUHuk9v/go-testutil"
	pstore "gx/ipfs/QmPiemjiKBC9VA7vZF82m4x1oygtg2c2YVqag8PX7dN1BD/go-libp2p-peerstore"
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
