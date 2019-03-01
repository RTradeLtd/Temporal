package pstoreds

import (
	"context"
	"time"

	base32 "gx/ipfs/QmfVj3x4D6Jkq9SEoi5n2NmoUomLwoeiwnYz2KQa15wRw6/base32"

	ds "gx/ipfs/QmUadX5EcvrBmxAV9sE7wUWtWSqxns5K84qKJBixmcT1w9/go-datastore"
	query "gx/ipfs/QmUadX5EcvrBmxAV9sE7wUWtWSqxns5K84qKJBixmcT1w9/go-datastore/query"

	peer "gx/ipfs/QmYVXrKrKHDC9FobgmcmshCDyWwdrfwfanNQN4oxJ9Fk3h/go-libp2p-peer"
	pstore "gx/ipfs/QmaCTz9RkrU13bm9kMB54f7atgqM4qkjDZpRwRoJiWXEqs/go-libp2p-peerstore"
)

// Configuration object for the peerstore.
type Options struct {
	// The size of the in-memory cache. A value of 0 or lower disables the cache.
	CacheSize uint

	// Sweep interval to purge expired addresses from the datastore. If this is a zero value, GC will not run
	// automatically, but it'll be available on demand via explicit calls.
	GCPurgeInterval time.Duration

	// Interval to renew the GC lookahead window. If this is a zero value, lookahead will be disabled and we'll
	// traverse the entire datastore for every purge cycle.
	GCLookaheadInterval time.Duration

	// Initial delay before GC processes start. Intended to give the system breathing room to fully boot
	// before starting GC.
	GCInitialDelay time.Duration
}

// DefaultOpts returns the default options for a persistent peerstore, with the full-purge GC algorithm:
//
// * Cache size: 1024.
// * GC purge interval: 2 hours.
// * GC lookahead interval: disabled.
// * GC initial delay: 60 seconds.
func DefaultOpts() Options {
	return Options{
		CacheSize:           1024,
		GCPurgeInterval:     2 * time.Hour,
		GCLookaheadInterval: 0,
		GCInitialDelay:      60 * time.Second,
	}
}

// NewPeerstore creates a peerstore backed by the provided persistent datastore.
func NewPeerstore(ctx context.Context, store ds.Batching, opts Options) (pstore.Peerstore, error) {
	addrBook, err := NewAddrBook(ctx, store, opts)
	if err != nil {
		return nil, err
	}

	keyBook, err := NewKeyBook(ctx, store, opts)
	if err != nil {
		return nil, err
	}

	peerMetadata, err := NewPeerMetadata(ctx, store, opts)
	if err != nil {
		return nil, err
	}

	ps := pstore.NewPeerstore(keyBook, addrBook, peerMetadata)
	return ps, nil
}

// uniquePeerIds extracts and returns unique peer IDs from database keys.
func uniquePeerIds(ds ds.Datastore, prefix ds.Key, extractor func(result query.Result) string) (peer.IDSlice, error) {
	var (
		q       = query.Query{Prefix: prefix.String(), KeysOnly: true}
		results query.Results
		err     error
	)

	if results, err = ds.Query(q); err != nil {
		log.Error(err)
		return nil, err
	}

	defer results.Close()

	idset := make(map[string]struct{})
	for result := range results.Next() {
		k := extractor(result)
		idset[k] = struct{}{}
	}

	if len(idset) == 0 {
		return peer.IDSlice{}, nil
	}

	ids := make(peer.IDSlice, 0, len(idset))
	for id := range idset {
		pid, _ := base32.RawStdEncoding.DecodeString(id)
		id, _ := peer.IDFromBytes(pid)
		ids = append(ids, id)
	}
	return ids, nil
}
