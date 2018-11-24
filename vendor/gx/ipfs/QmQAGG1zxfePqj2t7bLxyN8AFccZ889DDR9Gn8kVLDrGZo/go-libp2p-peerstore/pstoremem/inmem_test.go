package pstoremem

import (
	"testing"

	pstore "gx/ipfs/QmQAGG1zxfePqj2t7bLxyN8AFccZ889DDR9Gn8kVLDrGZo/go-libp2p-peerstore"
	pt "gx/ipfs/QmQAGG1zxfePqj2t7bLxyN8AFccZ889DDR9Gn8kVLDrGZo/go-libp2p-peerstore/test"
)

func TestInMemoryPeerstore(t *testing.T) {
	pt.TestPeerstore(t, func() (pstore.Peerstore, func()) {
		return NewPeerstore(), nil
	})
}

func TestInMemoryAddrBook(t *testing.T) {
	pt.TestAddrBook(t, func() (pstore.AddrBook, func()) {
		return NewAddrBook(), nil
	})
}

func TestInMemoryKeyBook(t *testing.T) {
	pt.TestKeyBook(t, func() (pstore.KeyBook, func()) {
		return NewKeyBook(), nil
	})
}

func BenchmarkInMemoryPeerstore(b *testing.B) {
	pt.BenchmarkPeerstore(b, func() (pstore.Peerstore, func()) {
		return NewPeerstore(), nil
	}, "InMem")
}
