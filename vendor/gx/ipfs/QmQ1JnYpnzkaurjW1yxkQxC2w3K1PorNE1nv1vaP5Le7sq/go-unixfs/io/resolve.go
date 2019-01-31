package io

import (
	"context"

	ft "gx/ipfs/QmQ1JnYpnzkaurjW1yxkQxC2w3K1PorNE1nv1vaP5Le7sq/go-unixfs"
	hamt "gx/ipfs/QmQ1JnYpnzkaurjW1yxkQxC2w3K1PorNE1nv1vaP5Le7sq/go-unixfs/hamt"
	dag "gx/ipfs/Qmb2UEG2TAeVrEJSjqsZF7Y2he7wRDkrdt6c3bECxwZf4k/go-merkledag"

	ipld "gx/ipfs/QmRL22E4paat7ky7vx9MLpR97JHHbFPrg3ytFQw6qp1y1s/go-ipld-format"
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
