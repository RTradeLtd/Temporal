package bstest

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	. "gx/ipfs/QmYPZzd9VqmJDwxUnThfeSbV1Y5o53aVPDijTB7j7rS9Ep/go-blockservice"

	u "gx/ipfs/QmNohiVssaPw3KVLZik59DBVGTSm2dGvYT9eoXt5DQ36Yz/go-ipfs-util"
	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	blockstore "gx/ipfs/QmS2aqUZLJp8kF1ihE5rvDGE5LvmKDPnx32w9Z1BW9xLV5/go-ipfs-blockstore"
	blocks "gx/ipfs/QmWoXtvgC8inqFkAATB7cp2Dax7XBi9VDvSg9RCCZufmRk/go-block-format"
	offline "gx/ipfs/QmYZwey1thDTynSrvd6qQkX24UpTka6TFhQ2v569UpoqxD/go-ipfs-exchange-offline"
	ds "gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore"
	dssync "gx/ipfs/Qmf4xQhNomPNhrtZc67qSnfJSjxjXs9LWvknJtSXwimPrM/go-datastore/sync"
)

func newObject(data []byte) blocks.Block {
	return blocks.NewBlock(data)
}

func TestBlocks(t *testing.T) {
	bstore := blockstore.NewBlockstore(dssync.MutexWrap(ds.NewMapDatastore()))
	bs := New(bstore, offline.Exchange(bstore))
	defer bs.Close()

	o := newObject([]byte("beep boop"))
	h := cid.NewCidV0(u.Hash([]byte("beep boop")))
	if !o.Cid().Equals(h) {
		t.Error("Block key and data multihash key not equal")
	}

	err := bs.AddBlock(o)
	if err != nil {
		t.Error("failed to add block to BlockService", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	b2, err := bs.GetBlock(ctx, o.Cid())
	if err != nil {
		t.Error("failed to retrieve block from BlockService", err)
		return
	}

	if !o.Cid().Equals(b2.Cid()) {
		t.Error("Block keys not equal.")
	}

	if !bytes.Equal(o.RawData(), b2.RawData()) {
		t.Error("Block data is not equal.")
	}
}

func makeObjects(n int) []blocks.Block {
	var out []blocks.Block
	for i := 0; i < n; i++ {
		out = append(out, newObject([]byte(fmt.Sprintf("object %d", i))))
	}
	return out
}

func TestGetBlocksSequential(t *testing.T) {
	var servs = Mocks(4)
	for _, s := range servs {
		defer s.Close()
	}
	objs := makeObjects(50)

	var cids []cid.Cid
	for _, o := range objs {
		cids = append(cids, o.Cid())
		servs[0].AddBlock(o)
	}

	t.Log("one instance at a time, get blocks concurrently")

	for i := 1; i < len(servs); i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
		defer cancel()
		out := servs[i].GetBlocks(ctx, cids)
		gotten := make(map[string]blocks.Block)
		for blk := range out {
			if _, ok := gotten[blk.Cid().KeyString()]; ok {
				t.Fatal("Got duplicate block!")
			}
			gotten[blk.Cid().KeyString()] = blk
		}
		if len(gotten) != len(objs) {
			t.Fatalf("Didnt get enough blocks back: %d/%d", len(gotten), len(objs))
		}
	}
}
