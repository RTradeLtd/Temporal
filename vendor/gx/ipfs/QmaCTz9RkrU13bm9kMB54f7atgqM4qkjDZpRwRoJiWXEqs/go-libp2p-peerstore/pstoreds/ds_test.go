package pstoreds

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	badger "gx/ipfs/QmTNJogwkhnbHeRmAXWtzvN2KgVko2oNmHHQN1ggHVhF91/go-ds-badger"
	ds "gx/ipfs/QmUadX5EcvrBmxAV9sE7wUWtWSqxns5K84qKJBixmcT1w9/go-datastore"
	leveldb "gx/ipfs/QmbgYmpUkuCDnXi4hci3Jt797iVXbpuBKRTCqGz57h48Sk/go-ds-leveldb"

	pstore "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore"
	pt "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore/test"
)

type datastoreFactory func(tb testing.TB) (ds.Batching, func())

var dstores = map[string]datastoreFactory{
	"Badger": badgerStore,
	// "Leveldb": leveldbStore,
}

func TestDsPeerstore(t *testing.T) {
	for name, dsFactory := range dstores {
		t.Run(name, func(t *testing.T) {
			pt.TestPeerstore(t, peerstoreFactory(t, dsFactory, DefaultOpts()))
		})
	}
}

func TestDsAddrBook(t *testing.T) {
	for name, dsFactory := range dstores {
		t.Run(name+" Cacheful", func(t *testing.T) {
			t.Parallel()

			opts := DefaultOpts()
			opts.GCPurgeInterval = 1 * time.Second
			opts.CacheSize = 1024

			pt.TestAddrBook(t, addressBookFactory(t, dsFactory, opts))
		})

		t.Run(name+" Cacheless", func(t *testing.T) {
			t.Parallel()

			opts := DefaultOpts()
			opts.GCPurgeInterval = 1 * time.Second
			opts.CacheSize = 0

			pt.TestAddrBook(t, addressBookFactory(t, dsFactory, opts))
		})
	}
}

func TestDsKeyBook(t *testing.T) {
	for name, dsFactory := range dstores {
		t.Run(name, func(t *testing.T) {
			pt.TestKeyBook(t, keyBookFactory(t, dsFactory, DefaultOpts()))
		})
	}
}

func BenchmarkDsKeyBook(b *testing.B) {
	for name, dsFactory := range dstores {
		b.Run(name, func(b *testing.B) {
			pt.BenchmarkKeyBook(b, keyBookFactory(b, dsFactory, DefaultOpts()))
		})
	}
}

func BenchmarkDsPeerstore(b *testing.B) {
	caching := DefaultOpts()
	caching.CacheSize = 1024

	cacheless := DefaultOpts()
	cacheless.CacheSize = 0

	for name, dsFactory := range dstores {
		b.Run(name, func(b *testing.B) {
			pt.BenchmarkPeerstore(b, peerstoreFactory(b, dsFactory, caching), "Caching")
		})
		b.Run(name, func(b *testing.B) {
			pt.BenchmarkPeerstore(b, peerstoreFactory(b, dsFactory, cacheless), "Cacheless")
		})
	}
}

func badgerStore(tb testing.TB) (ds.Batching, func()) {
	dataPath, err := ioutil.TempDir(os.TempDir(), "badger")
	if err != nil {
		tb.Fatal(err)
	}
	store, err := badger.NewDatastore(dataPath, nil)
	if err != nil {
		tb.Fatal(err)
	}
	closer := func() {
		store.Close()
		os.RemoveAll(dataPath)
	}
	return store, closer
}

func leveldbStore(tb testing.TB) (ds.TxnDatastore, func()) {
	dataPath, err := ioutil.TempDir(os.TempDir(), "leveldb")
	if err != nil {
		tb.Fatal(err)
	}
	store, err := leveldb.NewDatastore(dataPath, nil)
	if err != nil {
		tb.Fatal(err)
	}
	closer := func() {
		store.Close()
		os.RemoveAll(dataPath)
	}
	return store, closer
}

func peerstoreFactory(tb testing.TB, storeFactory datastoreFactory, opts Options) pt.PeerstoreFactory {
	return func() (pstore.Peerstore, func()) {
		store, storeCloseFn := storeFactory(tb)
		ps, err := NewPeerstore(context.Background(), store, opts)
		if err != nil {
			tb.Fatal(err)
		}
		closer := func() {
			ps.Close()
			storeCloseFn()
		}
		return ps, closer
	}
}

func addressBookFactory(tb testing.TB, storeFactory datastoreFactory, opts Options) pt.AddrBookFactory {
	return func() (pstore.AddrBook, func()) {
		store, closeFunc := storeFactory(tb)
		ab, err := NewAddrBook(context.Background(), store, opts)
		if err != nil {
			tb.Fatal(err)
		}
		closer := func() {
			ab.Close()
			closeFunc()
		}
		return ab, closer
	}
}

func keyBookFactory(tb testing.TB, storeFactory datastoreFactory, opts Options) pt.KeyBookFactory {
	return func() (pstore.KeyBook, func()) {
		store, storeCloseFn := storeFactory(tb)
		kb, err := NewKeyBook(context.Background(), store, opts)
		if err != nil {
			tb.Fatal(err)
		}
		return kb, storeCloseFn
	}
}
