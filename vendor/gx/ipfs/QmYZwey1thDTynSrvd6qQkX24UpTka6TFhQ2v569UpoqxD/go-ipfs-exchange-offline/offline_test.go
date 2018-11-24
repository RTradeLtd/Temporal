package offline

import (
	"context"
	"testing"

	u "gx/ipfs/QmNohiVssaPw3KVLZik59DBVGTSm2dGvYT9eoXt5DQ36Yz/go-ipfs-util"
	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	blockstore "gx/ipfs/QmS2aqUZLJp8kF1ihE5rvDGE5LvmKDPnx32w9Z1BW9xLV5/go-ipfs-blockstore"
	blocksutil "gx/ipfs/QmWTtpEozefF75GPw8pfsjdK12a6hZSW4CrzeecXbsVzek/go-ipfs-blocksutil"
	blocks "gx/ipfs/QmWoXtvgC8inqFkAATB7cp2Dax7XBi9VDvSg9RCCZufmRk/go-block-format"
	ds "gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore"
	ds_sync "gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore/sync"
)

func TestBlockReturnsErr(t *testing.T) {
	off := Exchange(bstore())
	c := cid.NewCidV0(u.Hash([]byte("foo")))
	_, err := off.GetBlock(context.Background(), c)
	if err != nil {
		return // as desired
	}
	t.Fail()
}

func TestHasBlockReturnsNil(t *testing.T) {
	store := bstore()
	ex := Exchange(store)
	block := blocks.NewBlock([]byte("data"))

	err := ex.HasBlock(block)
	if err != nil {
		t.Fail()
	}

	if _, err := store.Get(block.Cid()); err != nil {
		t.Fatal(err)
	}
}

func TestGetBlocks(t *testing.T) {
	store := bstore()
	ex := Exchange(store)
	g := blocksutil.NewBlockGenerator()

	expected := g.Blocks(2)

	for _, b := range expected {
		if err := ex.HasBlock(b); err != nil {
			t.Fail()
		}
	}

	request := func() []cid.Cid {
		var ks []cid.Cid

		for _, b := range expected {
			ks = append(ks, b.Cid())
		}
		return ks
	}()

	received, err := ex.GetBlocks(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}

	var count int
	for range received {
		count++
	}
	if len(expected) != count {
		t.Fail()
	}
}

func bstore() blockstore.Blockstore {
	return blockstore.NewBlockstore(ds_sync.MutexWrap(ds.NewMapDatastore()))
}
