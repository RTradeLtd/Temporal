package utils

import (
	offline "gx/ipfs/QmS6mo1dPpHdYsVkm27BRZDLxpKBCiJKUH8fHX15XFfMez/go-ipfs-exchange-offline"
	blockstore "gx/ipfs/QmadMhXJLHMFjpRmh85XjpmVDkEtQpNYEZNRpWRvYVLrvb/go-ipfs-blockstore"
	datastore "gx/ipfs/QmeiCcJfDW1GJnWUArudsv5rQsihpi4oyddPhdqo3CfX6i/go-datastore"
	"io"

	"github.com/ipfs/go-ipfs/blockservice"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreunix"
	"github.com/ipfs/go-ipfs/merkledag"
	"github.com/ipfs/go-ipfs/pin"
)

// GenerateIpfsMultiHashForFile is used to calculate an IPFS multihash
// for a given file, without actually broadcasting it to the network
// or even adding it at all. It was graciously adopted thanks to @hinshun
// over on discuss.ipfs.io. Overtime we willallow calculating different hash types
func GenerateIpfsMultiHashForFile(r io.Reader) (string, error) {
	dstore := datastore.NewMapDatastore()
	bstore := blockstore.NewBlockstore(dstore)
	bserv := blockservice.New(bstore, offline.Exchange(bstore))
	dserv := merkledag.NewDAGService(bserv)
	n := &core.IpfsNode{
		Blockstore: blockstore.NewGCBlockstore(bstore, blockstore.NewGCLocker()),
		Pinning:    pin.NewPinner(dstore, dserv, dserv),
		DAG:        dserv,
	}
	return coreunix.Add(n, r)
}
