package utils

import (
	"io"

	"github.com/ipsn/go-ipfs/core"
	"github.com/ipsn/go-ipfs/core/coreunix"
	"github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-blockservice"
	datastore "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-datastore"
	blockstore "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-ipfs-blockstore"
	offline "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-ipfs-exchange-offline"
	"github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-merkledag"
	"github.com/ipsn/go-ipfs/pin"
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
