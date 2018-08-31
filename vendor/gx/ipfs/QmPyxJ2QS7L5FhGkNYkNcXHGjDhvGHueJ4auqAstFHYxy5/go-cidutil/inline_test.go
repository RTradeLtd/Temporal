package cidutil

import (
	"math/rand"
	"testing"

	mhash "gx/ipfs/QmPnFwZ2JXKnXgMw8CdBPxn7FWh6LLdjUjxV1fKHuJnkr8/go-multihash"
	cid "gx/ipfs/QmZFbDTY9jfSBms2MchvYM9oYRbAF19K7Pby47yDBfpPrb/go-cid"
)

func TestInlineBuilderSmallValue(t *testing.T) {
	builder := InlineBuilder{cid.V0Builder{}, 64}
	c, err := builder.Sum([]byte("Hello World"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Prefix().MhType != mhash.ID {
		t.Fatal("Inliner builder failed to use ID Multihash on small values")
	}
}

func TestInlinerBuilderLargeValue(t *testing.T) {
	builder := InlineBuilder{cid.V0Builder{}, 64}
	data := make([]byte, 512)
	rand.Read(data)
	c, err := builder.Sum(data)
	if err != nil {
		t.Fatal(err)
	}
	if c.Prefix().MhType == mhash.ID {
		t.Fatal("Inliner builder used ID Multihash on large values")
	}
}
