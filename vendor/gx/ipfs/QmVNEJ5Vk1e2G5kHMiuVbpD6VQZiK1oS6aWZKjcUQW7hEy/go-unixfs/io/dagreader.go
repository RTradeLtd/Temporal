package io

import (
	"context"
	"errors"
	"io"

	mdag "gx/ipfs/QmRDaC5z6yXkXTTSWzaxs2sSVBon5RRCN6eNtMmpuHtKCr/go-merkledag"
	ft "gx/ipfs/QmVNEJ5Vk1e2G5kHMiuVbpD6VQZiK1oS6aWZKjcUQW7hEy/go-unixfs"
	ftpb "gx/ipfs/QmVNEJ5Vk1e2G5kHMiuVbpD6VQZiK1oS6aWZKjcUQW7hEy/go-unixfs/pb"

	ipld "gx/ipfs/QmX5CsuHyVZeTLxgRSYkgLSDQKb9UjE8xnhQzCEJWWWFsC/go-ipld-format"
)

// Common errors
var (
	ErrIsDir            = errors.New("this dag node is a directory")
	ErrCantReadSymlinks = errors.New("cannot currently read symlinks")
	ErrUnkownNodeType   = errors.New("unknown node type")
)

// A DagReader provides read-only read and seek acess to a unixfs file.
// Different implementations of readers are used for the different
// types of unixfs/protobuf-encoded nodes.
type DagReader interface {
	ReadSeekCloser
	Size() uint64
	CtxReadFull(context.Context, []byte) (int, error)
}

// A ReadSeekCloser implements interfaces to read, copy, seek and close.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
	io.WriterTo
}

// NewDagReader creates a new reader object that reads the data represented by
// the given node, using the passed in DAGService for data retrieval
func NewDagReader(ctx context.Context, n ipld.Node, serv ipld.NodeGetter) (DagReader, error) {
	switch n := n.(type) {
	case *mdag.RawNode:
		return NewBufDagReader(n.RawData()), nil
	case *mdag.ProtoNode:
		fsNode, err := ft.FSNodeFromBytes(n.Data())
		if err != nil {
			return nil, err
		}

		switch fsNode.Type() {
		case ftpb.Data_Directory, ftpb.Data_HAMTShard:
			// Dont allow reading directories
			return nil, ErrIsDir
		case ftpb.Data_File, ftpb.Data_Raw:
			return NewPBFileReader(ctx, n, fsNode, serv), nil
		case ftpb.Data_Metadata:
			if len(n.Links()) == 0 {
				return nil, errors.New("incorrectly formatted metadata object")
			}
			child, err := n.Links()[0].GetNode(ctx, serv)
			if err != nil {
				return nil, err
			}

			childpb, ok := child.(*mdag.ProtoNode)
			if !ok {
				return nil, mdag.ErrNotProtobuf
			}
			return NewDagReader(ctx, childpb, serv)
		case ftpb.Data_Symlink:
			return nil, ErrCantReadSymlinks
		default:
			return nil, ft.ErrUnrecognizedType
		}
	default:
		return nil, ErrUnkownNodeType
	}
}
