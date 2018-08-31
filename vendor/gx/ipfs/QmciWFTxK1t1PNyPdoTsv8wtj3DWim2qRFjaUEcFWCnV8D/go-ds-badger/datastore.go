package badger

import (
	"fmt"
	"strings"

	osh "gx/ipfs/QmXuBJ7DR6k3rmUEKtvVMhwjmXDuJgXXPUt4LQXKBMsU93/go-os-helper"
	badger "gx/ipfs/Qmb3GBFCHMuzmi9EpH3pxpYBiviyc3tEPyDQHxZJQJSxj9/badger"

	goprocess "gx/ipfs/QmSF8fPo3jgVBAy8fpdjjYqgG87dkJgUprRBHRd2tmfgpP/goprocess"
	ds "gx/ipfs/QmVG5gxteQNEMhrS8prJSmU2C9rebtFuTd3SYZ5kE3YZ5k/go-datastore"
	dsq "gx/ipfs/QmVG5gxteQNEMhrS8prJSmU2C9rebtFuTd3SYZ5kE3YZ5k/go-datastore/query"
)

type Datastore struct {
	DB *badger.DB

	gcDiscardRatio float64
}

// Options are the badger datastore options, reexported here for convenience.
type Options struct {
	gcDiscardRatio float64

	badger.Options
}

var DefaultOptions = Options{
	gcDiscardRatio: 0.1,

	Options: badger.DefaultOptions,
}

// NewDatastore creates a new badger datastore.
//
// DO NOT set the Dir and/or ValuePath fields of opt, they will be set for you.
func NewDatastore(path string, options *Options) (*Datastore, error) {
	// Copy the options because we modify them.
	var opt badger.Options
	var gcDiscardRatio float64
	if options == nil {
		opt = badger.DefaultOptions
		gcDiscardRatio = DefaultOptions.gcDiscardRatio
	} else {
		opt = options.Options
		gcDiscardRatio = options.gcDiscardRatio
	}

	if osh.IsWindows() && opt.SyncWrites {
		opt.Truncate = true
	}

	opt.Dir = path
	opt.ValueDir = path

	kv, err := badger.Open(opt)
	if err != nil {
		if strings.HasPrefix(err.Error(), "manifest has unsupported version:") {
			err = fmt.Errorf("unsupported badger version, use github.com/ipfs/badgerds-upgrade to upgrade: %s", err.Error())
		}
		return nil, err
	}

	return &Datastore{
		DB: kv,

		gcDiscardRatio: gcDiscardRatio,
	}, nil
}

func (d *Datastore) Put(key ds.Key, value []byte) error {
	txn := d.DB.NewTransaction(true)
	defer txn.Discard()

	err := txn.Set(key.Bytes(), value)
	if err != nil {
		return err
	}

	//TODO: Setting callback may potentially make this faster
	return txn.Commit(nil)
}

func (d *Datastore) Get(key ds.Key) (value []byte, err error) {
	txn := d.DB.NewTransaction(false)
	defer txn.Discard()

	item, err := txn.Get(key.Bytes())
	if err == badger.ErrKeyNotFound {
		err = ds.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	val, err := item.Value()
	if err != nil {
		return nil, err
	}

	out := make([]byte, len(val))
	copy(out, val)
	return out, nil
}

func (d *Datastore) Has(key ds.Key) (bool, error) {
	txn := d.DB.NewTransaction(false)
	defer txn.Discard()
	_, err := txn.Get(key.Bytes())
	if err == badger.ErrKeyNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func (d *Datastore) Delete(key ds.Key) error {
	txn := d.DB.NewTransaction(true)
	defer txn.Discard()
	err := txn.Delete(key.Bytes())
	if err != nil {
		return err
	}

	//TODO: callback may potentially make this faster
	return txn.Commit(nil)
}

func (d *Datastore) Query(q dsq.Query) (dsq.Results, error) {
	return d.QueryNew(q)
}

func (d *Datastore) QueryNew(q dsq.Query) (dsq.Results, error) {
	prefix := []byte(q.Prefix)
	opt := badger.DefaultIteratorOptions
	opt.PrefetchValues = !q.KeysOnly

	txn := d.DB.NewTransaction(false)

	it := txn.NewIterator(opt)
	it.Seek([]byte(q.Prefix))
	if q.Offset > 0 {
		for j := 0; j < q.Offset; j++ {
			it.Next()
		}
	}

	qrb := dsq.NewResultBuilder(q)

	qrb.Process.Go(func(worker goprocess.Process) {
		defer txn.Discard()
		defer it.Close()

		for sent := 0; it.ValidForPrefix(prefix); sent++ {
			if qrb.Query.Limit > 0 && sent >= qrb.Query.Limit {
				break
			}

			item := it.Item()

			k := string(item.Key())
			e := dsq.Entry{Key: k}

			var result dsq.Result
			if !q.KeysOnly {
				b, err := item.Value()
				if err != nil {
					result = dsq.Result{Error: err}
				} else {
					bytes := make([]byte, len(b))
					copy(bytes, b)
					e.Value = bytes
					result = dsq.Result{Entry: e}
				}
			} else {
				result = dsq.Result{Entry: e}
			}

			select {
			case qrb.Output <- result:
			case <-worker.Closing(): // client told us to close early
				return
			}
			it.Next()
		}

		return
	})

	go qrb.Process.CloseAfterChildren()

	// Now, apply remaining things (filters, order)
	qr := qrb.Results()
	for _, f := range q.Filters {
		qr = dsq.NaiveFilter(qr, f)
	}
	for _, o := range q.Orders {
		qr = dsq.NaiveOrder(qr, o)
	}

	return qr, nil
}

// DiskUsage implements the PersistentDatastore interface.
// It returns the sum of lsm and value log files sizes in bytes.
func (d *Datastore) DiskUsage() (uint64, error) {
	lsm, vlog := d.DB.Size()
	return uint64(lsm + vlog), nil
}

func (d *Datastore) Close() error {
	return d.DB.Close()
}

func (d *Datastore) IsThreadSafe() {}

type badgerBatch struct {
	db  *badger.DB
	txn *badger.Txn
}

func (d *Datastore) Batch() (ds.Batch, error) {
	return &badgerBatch{
		db:  d.DB,
		txn: d.DB.NewTransaction(true),
	}, nil
}

func (b *badgerBatch) Put(key ds.Key, value []byte) error {
	err := b.txn.Set(key.Bytes(), value)
	if err != nil {
		b.txn.Discard()
	}
	return err
}

func (b *badgerBatch) Delete(key ds.Key) error {
	err := b.txn.Delete(key.Bytes())
	if err != nil {
		b.txn.Discard()
	}
	return err
}

func (b *badgerBatch) Commit() error {
	//TODO: Setting callback may potentially make this faster
	return b.txn.Commit(nil)
}

func (d *Datastore) CollectGarbage() error {
	err := d.DB.RunValueLogGC(d.gcDiscardRatio)
	if err == badger.ErrNoRewrite {
		err = nil
	}
	return err
}
