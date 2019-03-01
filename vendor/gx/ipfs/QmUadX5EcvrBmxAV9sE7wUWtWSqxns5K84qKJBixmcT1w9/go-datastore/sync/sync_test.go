package sync

import (
	"testing"

	ds "gx/ipfs/QmUadX5EcvrBmxAV9sE7wUWtWSqxns5K84qKJBixmcT1w9/go-datastore"
	dstest "gx/ipfs/QmUadX5EcvrBmxAV9sE7wUWtWSqxns5K84qKJBixmcT1w9/go-datastore/test"
)

func TestSync(t *testing.T) {
	dstest.SubtestAll(t, MutexWrap(ds.NewMapDatastore()))
}
