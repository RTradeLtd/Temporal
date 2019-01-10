package test

import (
	"sort"
	"testing"

	ic "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"
	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"
	pt "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer/test"

	pstore "gx/ipfs/QmPiemjiKBC9VA7vZF82m4x1oygtg2c2YVqag8PX7dN1BD/go-libp2p-peerstore"
)

var keyBookSuite = map[string]func(kb pstore.KeyBook) func(*testing.T){
	"AddGetPrivKey":         testKeybookPrivKey,
	"AddGetPubKey":          testKeyBookPubKey,
	"PeersWithKeys":         testKeyBookPeers,
	"PubKeyAddedOnRetrieve": testInlinedPubKeyAddedOnRetrieve,
}

type KeyBookFactory func() (pstore.KeyBook, func())

func TestKeyBook(t *testing.T, factory KeyBookFactory) {
	for name, test := range keyBookSuite {
		// Create a new peerstore.
		kb, closeFunc := factory()

		// Run the test.
		t.Run(name, test(kb))

		// Cleanup.
		if closeFunc != nil {
			closeFunc()
		}
	}
}

func testKeybookPrivKey(kb pstore.KeyBook) func(t *testing.T) {
	return func(t *testing.T) {
		if peers := kb.PeersWithKeys(); len(peers) > 0 {
			t.Error("expected peers to be empty on init")
		}

		priv, _, err := pt.RandTestKeyPair(512)
		if err != nil {
			t.Error(err)
		}

		id, err := peer.IDFromPrivateKey(priv)
		if err != nil {
			t.Error(err)
		}

		if res := kb.PrivKey(id); res != nil {
			t.Error("retrieving private key should have failed")
		}

		err = kb.AddPrivKey(id, priv)
		if err != nil {
			t.Error(err)
		}

		if res := kb.PrivKey(id); !priv.Equals(res) {
			t.Error("retrieved private key did not match stored private key")
		}

		if peers := kb.PeersWithKeys(); len(peers) != 1 || peers[0] != id {
			t.Error("list of peers did not include test peer")
		}
	}
}

func testKeyBookPubKey(kb pstore.KeyBook) func(t *testing.T) {
	return func(t *testing.T) {
		if peers := kb.PeersWithKeys(); len(peers) > 0 {
			t.Error("expected peers to be empty on init")
		}

		_, pub, err := pt.RandTestKeyPair(512)
		if err != nil {
			t.Error(err)
		}

		id, err := peer.IDFromPublicKey(pub)
		if err != nil {
			t.Error(err)
		}

		if res := kb.PubKey(id); res != nil {
			t.Error("retrieving public key should have failed")
		}

		err = kb.AddPubKey(id, pub)
		if err != nil {
			t.Error(err)
		}

		if res := kb.PubKey(id); !pub.Equals(res) {
			t.Error("retrieved public key did not match stored public key")
		}

		if peers := kb.PeersWithKeys(); len(peers) != 1 || peers[0] != id {
			t.Error("list of peers did not include test peer")
		}
	}
}

func testKeyBookPeers(kb pstore.KeyBook) func(t *testing.T) {
	return func(t *testing.T) {
		if peers := kb.PeersWithKeys(); len(peers) > 0 {
			t.Error("expected peers to be empty on init")
		}

		var peers peer.IDSlice
		for i := 0; i < 10; i++ {
			// Add a public key.
			_, pub, _ := pt.RandTestKeyPair(512)
			p1, _ := peer.IDFromPublicKey(pub)
			kb.AddPubKey(p1, pub)

			// Add a private key.
			priv, _, _ := pt.RandTestKeyPair(512)
			p2, _ := peer.IDFromPrivateKey(priv)
			kb.AddPrivKey(p2, priv)

			peers = append(peers, []peer.ID{p1, p2}...)
		}

		kbPeers := kb.PeersWithKeys()
		sort.Sort(kbPeers)
		sort.Sort(peers)

		for i, p := range kbPeers {
			if p != peers[i] {
				t.Errorf("mismatch of peer at index %d", i)
			}
		}
	}
}

func testInlinedPubKeyAddedOnRetrieve(kb pstore.KeyBook) func(t *testing.T) {
	return func(t *testing.T) {
		t.Skip("key inlining disabled for now: see libp2p/specs#111")

		if peers := kb.PeersWithKeys(); len(peers) > 0 {
			t.Error("expected peers to be empty on init")
		}

		// Key small enough for inlining.
		_, pub, err := ic.GenerateKeyPair(ic.Ed25519, 256)
		if err != nil {
			t.Error(err)
		}

		id, err := peer.IDFromPublicKey(pub)
		if err != nil {
			t.Error(err)
		}

		pubKey := kb.PubKey(id)
		if !pubKey.Equals(pub) {
			t.Error("mismatch between original public key and keybook-calculated one")
		}
	}
}

var keybookBenchmarkSuite = map[string]func(kb pstore.KeyBook) func(*testing.B){
	"PubKey":        benchmarkPubKey,
	"AddPubKey":     benchmarkAddPubKey,
	"PrivKey":       benchmarkPrivKey,
	"AddPrivKey":    benchmarkAddPrivKey,
	"PeersWithKeys": benchmarkPeersWithKeys,
}

func BenchmarkKeyBook(b *testing.B, factory KeyBookFactory) {
	ordernames := make([]string, 0, len(keybookBenchmarkSuite))
	for name := range keybookBenchmarkSuite {
		ordernames = append(ordernames, name)
	}
	sort.Strings(ordernames)
	for _, name := range ordernames {
		bench := keybookBenchmarkSuite[name]
		kb, closeFunc := factory()

		b.Run(name, bench(kb))

		if closeFunc != nil {
			closeFunc()
		}
	}
}

func benchmarkPubKey(kb pstore.KeyBook) func(*testing.B) {
	return func(b *testing.B) {
		_, pub, err := pt.RandTestKeyPair(512)
		if err != nil {
			b.Error(err)
		}

		id, err := peer.IDFromPublicKey(pub)
		if err != nil {
			b.Error(err)
		}

		err = kb.AddPubKey(id, pub)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			kb.PubKey(id)
		}
	}
}

func benchmarkAddPubKey(kb pstore.KeyBook) func(*testing.B) {
	return func(b *testing.B) {
		_, pub, err := pt.RandTestKeyPair(512)
		if err != nil {
			b.Error(err)
		}

		id, err := peer.IDFromPublicKey(pub)
		if err != nil {
			b.Error(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			kb.AddPubKey(id, pub)
		}
	}
}

func benchmarkPrivKey(kb pstore.KeyBook) func(*testing.B) {
	return func(b *testing.B) {
		priv, _, err := pt.RandTestKeyPair(512)
		if err != nil {
			b.Error(err)
		}

		id, err := peer.IDFromPrivateKey(priv)
		if err != nil {
			b.Error(err)
		}

		err = kb.AddPrivKey(id, priv)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			kb.PrivKey(id)
		}
	}
}

func benchmarkAddPrivKey(kb pstore.KeyBook) func(*testing.B) {
	return func(b *testing.B) {
		priv, _, err := pt.RandTestKeyPair(512)
		if err != nil {
			b.Error(err)
		}

		id, err := peer.IDFromPrivateKey(priv)
		if err != nil {
			b.Error(err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			kb.AddPrivKey(id, priv)
		}
	}
}

func benchmarkPeersWithKeys(kb pstore.KeyBook) func(*testing.B) {
	return func(b *testing.B) {
		for i := 0; i < 10; i++ {
			priv, pub, err := pt.RandTestKeyPair(512)
			if err != nil {
				b.Error(err)
			}

			id, err := peer.IDFromPublicKey(pub)
			if err != nil {
				b.Error(err)
			}

			err = kb.AddPubKey(id, pub)
			if err != nil {
				b.Fatal(err)
			}
			err = kb.AddPrivKey(id, priv)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			kb.PeersWithKeys()
		}
	}
}
