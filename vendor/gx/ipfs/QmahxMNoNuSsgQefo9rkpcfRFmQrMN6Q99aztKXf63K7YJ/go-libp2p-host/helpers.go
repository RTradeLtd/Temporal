package host

import pstore "gx/ipfs/QmQAGG1zxfePqj2t7bLxyN8AFccZ889DDR9Gn8kVLDrGZo/go-libp2p-peerstore"

// PeerInfoFromHost returns a PeerInfo struct with the Host's ID and all of its Addrs.
func PeerInfoFromHost(h Host) *pstore.PeerInfo {
	return &pstore.PeerInfo{
		ID:    h.ID(),
		Addrs: h.Addrs(),
	}
}
