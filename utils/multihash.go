package utils

import (
	merkledag "gx/ipfs/QmRiQCJZ91B7VNmLvA6sxzDuBJGSojS3uXHHVuNr3iueNZ/go-merkledag"
	blockservice "gx/ipfs/QmbSB9Uh3wVgmiCb1fAb8zuC3qAE6un4kd1jvatUurfAmB/go-blockservice"
	"io"

	blockstore "gx/ipfs/QmcmpX42gtDv1fz24kau4wjS9hfwWj5VexWBKgGnWzsyag/go-ipfs-blockstore"

	datastore "gx/ipfs/QmVG5gxteQNEMhrS8prJSmU2C9rebtFuTd3SYZ5kE3YZ5k/go-datastore"

	offline "gx/ipfs/QmZxjqR9Qgompju73kakSoUj3rbVndAzky3oCDiBNCxPs1/go-ipfs-exchange-offline"

	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/coreunix"
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
