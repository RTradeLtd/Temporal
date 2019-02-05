package merkledag

import (
	"context"

	ipld "gx/ipfs/QmRL22E4paat7ky7vx9MLpR97JHHbFPrg3ytFQw6qp1y1s/go-ipld-format"
)

// SessionMaker is an object that can generate a new fetching session.
type SessionMaker interface {
	Session(context.Context) ipld.NodeGetter
}

// NewSession returns a session backed NodeGetter if the given NodeGetter
// implements SessionMaker.
func NewSession(ctx context.Context, g ipld.NodeGetter) ipld.NodeGetter {
	if sm, ok := g.(SessionMaker); ok {
		return sm.Session(ctx)
	}
	return g
}
