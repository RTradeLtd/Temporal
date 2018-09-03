package io

import (
	"context"

	dag "gx/ipfs/QmRDaC5z6yXkXTTSWzaxs2sSVBon5RRCN6eNtMmpuHtKCr/go-merkledag"
	ft "gx/ipfs/QmVNEJ5Vk1e2G5kHMiuVbpD6VQZiK1oS6aWZKjcUQW7hEy/go-unixfs"
	hamt "gx/ipfs/QmVNEJ5Vk1e2G5kHMiuVbpD6VQZiK1oS6aWZKjcUQW7hEy/go-unixfs/hamt"

	ipld "gx/ipfs/QmX5CsuHyVZeTLxgRSYkgLSDQKb9UjE8xnhQzCEJWWWFsC/go-ipld-format"
)

// ResolveUnixfsOnce resolves a single hop of a path through a graph in a
// unixfs context. This includes handling traversing sharded directories.
func ResolveUnixfsOnce(ctx context.Context, ds ipld.NodeGetter, nd ipld.Node, names []string) (*ipld.Link, []string, error) {
	switch nd := nd.(type) {
	case *dag.ProtoNode:
		upb, err := ft.FromBytes(nd.Data())
		if err != nil {
			// Not a unixfs node, use standard object traversal code
			lnk, err := nd.GetNodeLink(names[0])
			if err != nil {
				return nil, nil, err
			}

			return lnk, names[1:], nil
		}

		switch upb.GetType() {
		case ft.THAMTShard:
			rods := dag.NewReadOnlyDagService(ds)
			s, err := hamt.NewHamtFromDag(rods, nd)
			if err != nil {
				return nil, nil, err
			}

			out, err := s.Find(ctx, names[0])
			if err != nil {
				return nil, nil, err
			}

			return out, names[1:], nil
		default:
			lnk, err := nd.GetNodeLink(names[0])
			if err != nil {
				return nil, nil, err
			}

			return lnk, names[1:], nil
		}
	default:
		lnk, rest, err := nd.ResolveLink(names)
		if err != nil {
			return nil, nil, err
		}
		return lnk, rest, nil
	}
}
