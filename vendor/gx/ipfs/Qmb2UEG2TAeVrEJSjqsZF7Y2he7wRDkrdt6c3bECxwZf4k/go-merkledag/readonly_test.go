package merkledag_test

import (
	"context"
	"testing"

	. "gx/ipfs/Qmb2UEG2TAeVrEJSjqsZF7Y2he7wRDkrdt6c3bECxwZf4k/go-merkledag"
	dstest "gx/ipfs/Qmb2UEG2TAeVrEJSjqsZF7Y2he7wRDkrdt6c3bECxwZf4k/go-merkledag/test"

	cid "gx/ipfs/QmR8BauakNcBa3RbE4nbQu76PDiJgoQgz8AJdhJuiU4TAw/go-cid"
	ipld "gx/ipfs/QmRL22E4paat7ky7vx9MLpR97JHHbFPrg3ytFQw6qp1y1s/go-ipld-format"
)

func TestReadonlyProperties(t *testing.T) {
	ds := dstest.Mock()
	ro := NewReadOnlyDagService(ds)

	ctx := context.Background()
	nds := []ipld.Node{
		NewRawNode([]byte("foo1")),
		NewRawNode([]byte("foo2")),
		NewRawNode([]byte("foo3")),
		NewRawNode([]byte("foo4")),
	}
	cids := []cid.Cid{
		nds[0].Cid(),
		nds[1].Cid(),
		nds[2].Cid(),
		nds[3].Cid(),
	}

	// add to the actual underlying datastore
	if err := ds.Add(ctx, nds[2]); err != nil {
		t.Fatal(err)
	}
	if err := ds.Add(ctx, nds[3]); err != nil {
		t.Fatal(err)
	}

	if err := ro.Add(ctx, nds[0]); err != ErrReadOnly {
		t.Fatal("expected ErrReadOnly")
	}
	if err := ro.Add(ctx, nds[2]); err != ErrReadOnly {
		t.Fatal("expected ErrReadOnly")
	}

	if err := ro.AddMany(ctx, nds[0:1]); err != ErrReadOnly {
		t.Fatal("expected ErrReadOnly")
	}

	if err := ro.Remove(ctx, cids[3]); err != ErrReadOnly {
		t.Fatal("expected ErrReadOnly")
	}
	if err := ro.RemoveMany(ctx, cids[1:2]); err != ErrReadOnly {
		t.Fatal("expected ErrReadOnly")
	}

	if _, err := ro.Get(ctx, cids[0]); err != ipld.ErrNotFound {
		t.Fatal("expected ErrNotFound")
	}
	if _, err := ro.Get(ctx, cids[3]); err != nil {
		t.Fatal(err)
	}
}
