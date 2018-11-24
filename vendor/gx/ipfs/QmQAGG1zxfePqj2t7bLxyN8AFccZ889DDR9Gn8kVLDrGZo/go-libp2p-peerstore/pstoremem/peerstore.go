package pstoremem

import pstore "gx/ipfs/QmQAGG1zxfePqj2t7bLxyN8AFccZ889DDR9Gn8kVLDrGZo/go-libp2p-peerstore"

// NewPeerstore creates an in-memory threadsafe collection of peers.
func NewPeerstore() pstore.Peerstore {
	return pstore.NewPeerstore(
		NewKeyBook(),
		NewAddrBook(),
		NewPeerMetadata())
}
