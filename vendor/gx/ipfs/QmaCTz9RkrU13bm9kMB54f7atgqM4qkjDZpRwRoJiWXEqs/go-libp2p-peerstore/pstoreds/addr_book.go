package pstoreds

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	ds "gx/ipfs/QmUadX5EcvrBmxAV9sE7wUWtWSqxns5K84qKJBixmcT1w9/go-datastore"
	query "gx/ipfs/QmUadX5EcvrBmxAV9sE7wUWtWSqxns5K84qKJBixmcT1w9/go-datastore/query"
	logging "gx/ipfs/QmbkT7eMTyXfpeyB3ZMxxcxg7XH8t6uXp49jqzz4HB7BGF/go-log"

	peer "gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer"
	pstore "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore"
	pb "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore/pb"
	pstoremem "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore/pstoremem"

	lru "gx/ipfs/QmQjMHF8ptRgx4E57UFMiT4YM6kqaJeYxZ1MCDX23aw4rK/golang-lru"
	ma "gx/ipfs/QmTZBfrPJmjWsCvHEtX5FE6KimVJhsJg5sBbqEFYf4UZtL/go-multiaddr"
	b32 "gx/ipfs/QmfVj3x4D6Jkq9SEoi5n2NmoUomLwoeiwnYz2KQa15wRw6/base32"
)

type ttlWriteMode int

const (
	ttlOverride ttlWriteMode = iota
	ttlExtend
)

var (
	log = logging.Logger("peerstore/ds")

	// Peer addresses are stored db key pattern:
	// /peers/addrs/<b32 peer id no padding>
	addrBookBase = ds.NewKey("/peers/addrs")
)

// addrsRecord decorates the AddrBookRecord with locks and metadata.
type addrsRecord struct {
	sync.RWMutex
	*pb.AddrBookRecord
	dirty bool
}

// flush writes the record to the datastore by calling ds.Put, unless the record is
// marked for deletion, in which case we call ds.Delete. To be called within a lock.
func (r *addrsRecord) flush(write ds.Write) (err error) {
	key := addrBookBase.ChildString(b32.RawStdEncoding.EncodeToString([]byte(r.Id.ID)))
	if len(r.Addrs) == 0 {
		if err = write.Delete(key); err == nil {
			r.dirty = false
		}
		return err
	}

	data, err := r.Marshal()
	if err != nil {
		return err
	}
	if err = write.Put(key, data); err != nil {
		return err
	}
	// write succeeded; record is no longer dirty.
	r.dirty = false
	return nil
}

// clean is called on records to perform housekeeping. The return value indicates if the record was changed
// as a result of this call.
//
// clean does the following:
// * sorts addresses by expiration (soonest expiring first).
// * removes expired addresses.
//
// It short-circuits optimistically when there's nothing to do.
//
// clean is called from several points:
// * when accessing an entry.
// * when performing periodic GC.
// * after an entry has been modified (e.g. addresses have been added or removed, TTLs updated, etc.)
//
// If the return value is true, the caller should perform a flush immediately to sync the record with the store.
func (r *addrsRecord) clean() (chgd bool) {
	now := time.Now().Unix()
	if !r.dirty && len(r.Addrs) > 0 && r.Addrs[0].Expiry > now {
		// record is not dirty, and we have no expired entries to purge.
		return false
	}

	if len(r.Addrs) == 0 {
		// this is a ghost record; let's signal it has to be written.
		// flush() will take care of doing the deletion.
		return true
	}

	if r.dirty && len(r.Addrs) > 1 {
		// the record has been modified, so it may need resorting.
		// we keep addresses sorted by expiration, where 0 is the soonest expiring.
		sort.Slice(r.Addrs, func(i, j int) bool {
			return r.Addrs[i].Expiry < r.Addrs[j].Expiry
		})
	}

	// since addresses are sorted by expiration, we find the first
	// survivor and split the slice on its index.
	pivot := -1
	for i, addr := range r.Addrs {
		if addr.Expiry > now {
			break
		}
		pivot = i
	}

	r.Addrs = r.Addrs[pivot+1:]
	return r.dirty || pivot >= 0
}

// dsAddrBook is an address book backed by a Datastore with a GC procedure to purge expired entries. It uses an
// in-memory address stream manager. See the NewAddrBook for more information.
type dsAddrBook struct {
	ctx  context.Context
	opts Options

	cache       cache
	ds          ds.Batching
	gc          *dsAddrBookGc
	subsManager *pstoremem.AddrSubManager

	// controls children goroutine lifetime.
	childrenDone sync.WaitGroup
	cancelFn     func()
}

var _ pstore.AddrBook = (*dsAddrBook)(nil)

// NewAddrBook initializes a new datastore-backed address book. It serves as a drop-in replacement for pstoremem
// (memory-backed peerstore), and works with any datastore implementing the ds.Batching interface.
//
// Addresses and peer records are serialized into protobuf, storing one datastore entry per peer, along with metadata
// to control address expiration. To alleviate disk access and serde overhead, we internally use a read/write-through
// ARC cache, the size of which is adjustable via Options.CacheSize.
//
// The user has a choice of two GC algorithms:
//
//  - lookahead GC: minimises the amount of full store traversals by maintaining a time-indexed list of entries that
//    need to be visited within the period specified in Options.GCLookaheadInterval. This is useful in scenarios with
//    considerable TTL variance, coupled with datastores whose native iterators return entries in lexicographical key
//    order. Enable this mode by passing a value Options.GCLookaheadInterval > 0. Lookahead windows are jumpy, not
//    sliding. Purges operate exclusively over the lookahead window with periodicity Options.GCPurgeInterval.
//
//  - full-purge GC (default): performs a full visit of the store with periodicity Options.GCPurgeInterval. Useful when
//    the range of possible TTL values is small and the values themselves are also extreme, e.g. 10 minutes or
//    permanent, popular values used in other libp2p modules. In this cited case, optimizing with lookahead windows
//    makes little sense.
func NewAddrBook(ctx context.Context, store ds.Batching, opts Options) (ab *dsAddrBook, err error) {
	ctx, cancelFn := context.WithCancel(ctx)
	ab = &dsAddrBook{
		ctx:         ctx,
		ds:          store,
		opts:        opts,
		cancelFn:    cancelFn,
		subsManager: pstoremem.NewAddrSubManager(),
	}

	if opts.CacheSize > 0 {
		if ab.cache, err = lru.NewARC(int(opts.CacheSize)); err != nil {
			return nil, err
		}
	} else {
		ab.cache = new(noopCache)
	}

	if ab.gc, err = newAddressBookGc(ctx, ab); err != nil {
		return nil, err
	}

	return ab, nil
}

func (ab *dsAddrBook) Close() error {
	ab.cancelFn()
	ab.childrenDone.Wait()
	return nil
}

// loadRecord is a read-through fetch. It fetches a record from cache, falling back to the
// datastore upon a miss, and returning a newly initialized record if the peer doesn't exist.
//
// loadRecord calls clean() on an existing record before returning it. If the record changes
// as a result and the update argument is true, the resulting state is saved in the datastore.
//
// If the cache argument is true, the record is inserted in the cache when loaded from the datastore.
func (ab *dsAddrBook) loadRecord(id peer.ID, cache bool, update bool) (pr *addrsRecord, err error) {
	if e, ok := ab.cache.Get(id); ok {
		pr = e.(*addrsRecord)
		pr.Lock()
		defer pr.Unlock()

		if pr.clean() && update {
			err = pr.flush(ab.ds)
		}
		return pr, err
	}

	pr = &addrsRecord{AddrBookRecord: &pb.AddrBookRecord{}}
	key := addrBookBase.ChildString(b32.RawStdEncoding.EncodeToString([]byte(id)))
	data, err := ab.ds.Get(key)

	switch err {
	case ds.ErrNotFound:
		err = nil
		pr.Id = &pb.ProtoPeerID{ID: id}
	case nil:
		if err = pr.Unmarshal(data); err != nil {
			return nil, err
		}
		// this record is new and local for now (not in cache), so we don't need to lock.
		if pr.clean() && update {
			err = pr.flush(ab.ds)
		}
	default:
		return nil, err
	}

	if cache {
		ab.cache.Add(id, pr)
	}
	return pr, err
}

// AddAddr will add a new address if it's not already in the AddrBook.
func (ab *dsAddrBook) AddAddr(p peer.ID, addr ma.Multiaddr, ttl time.Duration) {
	ab.AddAddrs(p, []ma.Multiaddr{addr}, ttl)
}

// AddAddrs will add many new addresses if they're not already in the AddrBook.
func (ab *dsAddrBook) AddAddrs(p peer.ID, addrs []ma.Multiaddr, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	addrs = cleanAddrs(addrs)
	ab.setAddrs(p, addrs, ttl, ttlExtend)
}

// SetAddr will add or update the TTL of an address in the AddrBook.
func (ab *dsAddrBook) SetAddr(p peer.ID, addr ma.Multiaddr, ttl time.Duration) {
	ab.SetAddrs(p, []ma.Multiaddr{addr}, ttl)
}

// SetAddrs will add or update the TTLs of addresses in the AddrBook.
func (ab *dsAddrBook) SetAddrs(p peer.ID, addrs []ma.Multiaddr, ttl time.Duration) {
	addrs = cleanAddrs(addrs)
	if ttl <= 0 {
		ab.deleteAddrs(p, addrs)
		return
	}
	ab.setAddrs(p, addrs, ttl, ttlOverride)
}

// UpdateAddrs will update any addresses for a given peer and TTL combination to
// have a new TTL.
func (ab *dsAddrBook) UpdateAddrs(p peer.ID, oldTTL time.Duration, newTTL time.Duration) {
	pr, err := ab.loadRecord(p, true, false)
	if err != nil {
		log.Errorf("failed to update ttls for peer %s: %s\n", p.Pretty(), err)
	}

	pr.Lock()
	defer pr.Unlock()

	newExp := time.Now().Add(newTTL).Unix()
	for _, entry := range pr.Addrs {
		if entry.Ttl != int64(oldTTL) {
			continue
		}
		entry.Ttl, entry.Expiry = int64(newTTL), newExp
		pr.dirty = true
	}

	if pr.clean() {
		pr.flush(ab.ds)
	}
}

// Addrs returns all of the non-expired addresses for a given peer.
func (ab *dsAddrBook) Addrs(p peer.ID) []ma.Multiaddr {
	pr, err := ab.loadRecord(p, true, true)
	if err != nil {
		log.Warning("failed to load peerstore entry for peer %v while querying addrs, err: %v", p, err)
		return nil
	}

	pr.RLock()
	defer pr.RUnlock()

	addrs := make([]ma.Multiaddr, 0, len(pr.Addrs))
	for _, a := range pr.Addrs {
		addrs = append(addrs, a.Addr)
	}
	return addrs
}

// Peers returns all of the peer IDs for which the AddrBook has addresses.
func (ab *dsAddrBook) PeersWithAddrs() peer.IDSlice {
	ids, err := uniquePeerIds(ab.ds, addrBookBase, func(result query.Result) string {
		return ds.RawKey(result.Key).Name()
	})
	if err != nil {
		log.Errorf("error while retrieving peers with addresses: %v", err)
	}
	return ids
}

// AddrStream returns a channel on which all new addresses discovered for a
// given peer ID will be published.
func (ab *dsAddrBook) AddrStream(ctx context.Context, p peer.ID) <-chan ma.Multiaddr {
	initial := ab.Addrs(p)
	return ab.subsManager.AddrStream(ctx, p, initial)
}

// ClearAddrs will delete all known addresses for a peer ID.
func (ab *dsAddrBook) ClearAddrs(p peer.ID) {
	ab.cache.Remove(p)

	key := addrBookBase.ChildString(b32.RawStdEncoding.EncodeToString([]byte(p)))
	if err := ab.ds.Delete(key); err != nil {
		log.Errorf("failed to clear addresses for peer %s: %v", p.Pretty(), err)
	}
}

func (ab *dsAddrBook) setAddrs(p peer.ID, addrs []ma.Multiaddr, ttl time.Duration, mode ttlWriteMode) (err error) {
	pr, err := ab.loadRecord(p, true, false)
	if err != nil {
		return fmt.Errorf("failed to load peerstore entry for peer %v while setting addrs, err: %v", p, err)
	}

	pr.Lock()
	defer pr.Unlock()

	newExp := time.Now().Add(ttl).Unix()
	existed := make([]bool, len(addrs)) // keeps track of which addrs we found.

Outer:
	for i, incoming := range addrs {
		for _, have := range pr.Addrs {
			if incoming.Equal(have.Addr) {
				existed[i] = true
				if mode == ttlExtend && have.Expiry > newExp {
					// if we're only extending TTLs but the addr already has a longer one, we skip it.
					continue Outer
				}
				have.Expiry = newExp
				// we found the address, and addresses cannot be duplicate,
				// so let's move on to the next.
				continue Outer
			}
		}
	}

	// add addresses we didn't hold.
	var added []*pb.AddrBookRecord_AddrEntry
	for i, e := range existed {
		if e {
			continue
		}
		addr := addrs[i]
		entry := &pb.AddrBookRecord_AddrEntry{
			Addr:   &pb.ProtoAddr{Multiaddr: addr},
			Ttl:    int64(ttl),
			Expiry: newExp,
		}
		added = append(added, entry)
		// note: there's a minor chance that writing the record will fail, in which case we would've broadcast
		// the addresses without persisting them. This is very unlikely and not much of an issue.
		ab.subsManager.BroadcastAddr(p, addr)
	}

	pr.Addrs = append(pr.Addrs, added...)
	pr.dirty = true
	pr.clean()
	return pr.flush(ab.ds)
}

func (ab *dsAddrBook) deleteAddrs(p peer.ID, addrs []ma.Multiaddr) (err error) {
	pr, err := ab.loadRecord(p, false, false)
	if err != nil {
		return fmt.Errorf("failed to load peerstore entry for peer %v while deleting addrs, err: %v", p, err)
	}

	if pr.Addrs == nil {
		return nil
	}

	pr.Lock()
	defer pr.Unlock()

	// deletes addresses in place, and avoiding copies until we encounter the first deletion.
	survived := 0
	for i, addr := range pr.Addrs {
		for _, del := range addrs {
			if addr.Addr.Equal(del) {
				continue
			}
			if i != survived {
				pr.Addrs[survived] = pr.Addrs[i]
			}
			survived++
		}
	}
	pr.Addrs = pr.Addrs[:survived]

	pr.dirty = true
	pr.clean()
	return pr.flush(ab.ds)
}

func cleanAddrs(addrs []ma.Multiaddr) []ma.Multiaddr {
	clean := make([]ma.Multiaddr, 0, len(addrs))
	for _, addr := range addrs {
		if addr == nil {
			continue
		}
		clean = append(clean, addr)
	}
	return clean
}
