package format

import (
	"errors"
	"testing"

	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	blocks "gx/ipfs/QmWoXtvgC8inqFkAATB7cp2Dax7XBi9VDvSg9RCCZufmRk/go-block-format"
	mh "gx/ipfs/QmerPMzPk1mJVowm8KgmoknWa4yCYvvugMPsgWmDNUvDLW/go-multihash"
)

func init() {
	Register(cid.Raw, func(b blocks.Block) (Node, error) {
		node := &EmptyNode{}
		if b.RawData() != nil || !b.Cid().Equals(node.Cid()) {
			return nil, errors.New("can only decode empty blocks")
		}
		return node, nil
	})
}

func TestDecode(t *testing.T) {
	id, err := cid.Prefix{
		Version:  1,
		Codec:    cid.Raw,
		MhType:   mh.ID,
		MhLength: 0,
	}.Sum(nil)

	if err != nil {
		t.Fatalf("failed to create cid: %s", err)
	}

	block, err := blocks.NewBlockWithCid(nil, id)
	if err != nil {
		t.Fatalf("failed to create empty block: %s", err)
	}
	node, err := Decode(block)
	if err != nil {
		t.Fatalf("failed to decode empty node: %s", err)
	}
	if !node.Cid().Equals(id) {
		t.Fatalf("empty node doesn't have the right cid")
	}

	if _, ok := node.(*EmptyNode); !ok {
		t.Fatalf("empty node doesn't have the right type")
	}

}
