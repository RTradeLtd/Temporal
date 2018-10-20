// Package posinfo wraps offset information used by ipfs filestore nodes
package posinfo

import (
	"os"

	ipld "gx/ipfs/QmZtNq8dArGfnpCZfx2pUNY7UcjGhVp5qqwQ4hH6mpTMRQ/go-ipld-format"
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
