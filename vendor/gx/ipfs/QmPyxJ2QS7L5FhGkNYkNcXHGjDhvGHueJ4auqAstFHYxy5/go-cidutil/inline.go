package cidutil

import (
	mhash "gx/ipfs/QmPnFwZ2JXKnXgMw8CdBPxn7FWh6LLdjUjxV1fKHuJnkr8/go-multihash"
	cid "gx/ipfs/QmZFbDTY9jfSBms2MchvYM9oYRbAF19K7Pby47yDBfpPrb/go-cid"
)

// InlineBuilder is a cid.Builder that will use the id multihash when the
// size of the content is no more than limit
type InlineBuilder struct {
	cid.Builder     // Parent Builder
	Limit       int // Limit (inclusive)
}

// WithCodec implements the cid.Builder interface
func (p InlineBuilder) WithCodec(c uint64) cid.Builder {
	return InlineBuilder{p.Builder.WithCodec(c), p.Limit}
}

// Sum implements the cid.Builder interface
func (p InlineBuilder) Sum(data []byte) (*cid.Cid, error) {
	if len(data) > p.Limit {
		return p.Builder.Sum(data)
	}
	return cid.V1Builder{Codec: p.GetCodec(), MhType: mhash.ID}.Sum(data)
}
