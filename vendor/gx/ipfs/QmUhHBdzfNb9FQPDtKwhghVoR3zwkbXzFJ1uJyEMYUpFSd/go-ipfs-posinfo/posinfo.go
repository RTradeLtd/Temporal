// Package posinfo wraps offset information used by ipfs filestore nodes
package posinfo

import (
	"os"

	ipld "gx/ipfs/QmRL22E4paat7ky7vx9MLpR97JHHbFPrg3ytFQw6qp1y1s/go-ipld-format"
)

// PosInfo stores information about the file offset, its path and
// stat.
type PosInfo struct {
	Offset   uint64
	FullPath string
	Stat     os.FileInfo // can be nil
}

// FilestoreNode is an ipld.Node which arries PosInfo with it
// allowing to map it directly to a filesystem object.
type FilestoreNode struct {
	ipld.Node
	PosInfo *PosInfo
}
