package io

import (
	"context"

	ft "gx/ipfs/QmUnHNqhSB1JgzVCxL1Kz3yb4bdyB4q1Z9AD5AUBVmt3fZ/go-unixfs"
	hamt "gx/ipfs/QmUnHNqhSB1JgzVCxL1Kz3yb4bdyB4q1Z9AD5AUBVmt3fZ/go-unixfs/hamt"
	dag "gx/ipfs/QmcGt25mrjuB2kKW2zhPbXVZNHc4yoTDQ65NA8m6auP2f1/go-merkledag"

	ipld "gx/ipfs/QmcKKBwfz6FyQdHR2jsXrrF6XeSBXYL86anmWNewpFpoF5/go-ipld-format"
)

// ResolveUnixfsOnce resolves a single hop of a path through a graph in a
// unixfs context. This includes handling traversing sharded directories.
func ResolveUnixfsOnce(ctx context.Context, ds ipld.NodeGetter, nd ipld.Node, names []string) (*ipld.Link, []string, error) {
	pn, ok := nd.(*dag.ProtoNode)
	if ok {
		fsn, err := ft.FSNodeFromBytes(pn.Data())
		if err != nil {
			// Not a unixfs node, use standard object traversal code
			return nd.ResolveLink(names)
		}

		if fsn.Type() == ft.THAMTShard {
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
		}
	}

	return nd.ResolveLink(names)
}
