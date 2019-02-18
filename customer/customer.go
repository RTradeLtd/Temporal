package customer

import (
	"encoding/json"
	"sync"

	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/rtfs"
)

// NewManager is used to instantiate our customer object manager
func NewManager(um *models.UserManager, ipfs rtfs.Manager) *Manager {
	return &Manager{
		um:    um,
		ipfs:  ipfs,
		mutex: sync.Mutex{},
	}
}

// GetDeduplicatedStorageSpaceInBytes is used to get the deduplicated storage space used
// by this particular hash. It calculates deduplicated storage costs based on previous uploads
// the user has made, and does not consider upload from other users.
// it is a read-only operation, and returns the storage space that will be used by this upload
func (m *Manager) GetDeduplicatedStorageSpaceInBytes(username, hash string) (int, error) {
	// find the user model
	user, err := m.um.FindByUserName(username)
	if err != nil {
		return 0, err
	}
	// construct an empty object
	var object Object
	// unmarshal the customer object hash into our typed object
	if err := m.ipfs.DagGet(user.CustomerObjectHash, &object); err != nil {
		return 0, err
	}
	// if the customer object is empty, then we return full storage space
	// consumed by the given hash
	if len(object.UploadedRefs) == 0 && len(object.UploadedRootNodes) == 0 {
		// get the size of references
		size, _, err := rtfs.DedupAndCalculatePinSize(hash, m.ipfs)
		if err != nil {
			return 0, err
		}
		// get datasize of root node
		stats, err := m.ipfs.Stat(hash)
		if err != nil {
			return 0, err
		}
		// update size to include data size of root node
		size = size + int64(stats.DataSize)
		return int(size), nil
	}
	// search through the refs for the given hash
	refs, err := m.ipfs.Refs(hash, true, true)
	if err != nil {
		return 0, err
	}
	// refsToCalculate will hold all references that we need
	// to calculate th estorage size for
	var (
		// will hold all references we need to update+charge for
		refsToCalculate []string
		// will hold the total datasize we need to charge
		size int
		// will determine if we have updated the root hash
		rootUpdated bool
	)
	// iterate over all references for the requested hash
	// and mark whether or not a reference hasn't been seen before
	for _, ref := range refs {
		if !object.UploadedRefs[ref] {
			refsToCalculate = append(refsToCalculate, ref)
		}
	}
	// check if we need to calculate datasize of the roots
	if !object.UploadedRootNodes[hash] {
		stats, err := m.ipfs.Stat(hash)
		if err != nil {
			return 0, err
		}
		// update size
		size = stats.DataSize
		// mark as having updated the root
		rootUpdated = true
	}
	// if we have no refs to calculate size for, it means this upload
	// will consume no additional storage space, thus we can avoid charging them
	// all but the datasize of the root (if at all)
	if len(refsToCalculate) == 0 {
		// if we have updated the used roots of the customer object
		// this means we need to first update the customer object
		// stored in database
		if rootUpdated {
			// return the new customer object hash and the datasize
			return size, nil
		}
		// since we haven't updated the root, and we have no refs to charge for
		// we can return a 0 size
		return 0, nil
	}
	// calculate size of all references
	// also use this to update the refs of customer object
	for _, ref := range refsToCalculate {
		// get stats for the reference
		stats, err := m.ipfs.Stat(ref)
		if err != nil {
			return 0, err
		}
		// update datasize to include size of reference
		size = size + stats.DataSize
	}
	return size, nil
}

// Update is used to update a customers deduplicated storage object
func (m *Manager) Update(username, hash string) (string, error) {
	// we use mutex locks so that if a single user account makes two calls to the API in a row
	// we do not want a confused state of used roots and refs to occur.
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// find the user model
	user, err := m.um.FindByUserName(username)
	if err != nil {
		return "", err
	}
	// construct an empty object
	var object Object
	// unmarshal the customer object hash into our typed object
	if err := m.ipfs.DagGet(user.CustomerObjectHash, &object); err != nil {
		return "", err
	}
	// if the customer object is empty, then we need to update it with all references
	if len(object.UploadedRefs) == 0 && len(object.UploadedRootNodes) == 0 {
		object = Object{
			UploadedRefs:      make(map[string]bool),
			UploadedRootNodes: make(map[string]bool),
		}
		// get references for the requested hash
		refs, err := m.ipfs.Refs(hash, true, true)
		if err != nil {
			return "", err
		}
		// iterate over all references marking them as being used
		for _, v := range refs {
			object.UploadedRefs[v] = true
		}
		// mark the root node as being used
		object.UploadedRootNodes[hash] = true
		// update the customer object hash stored in our database
		resp, err := m.put(username, &object)
		if err != nil {
			return "", err
		}
		return resp, nil
	}
	// search through the refs for the given hash
	refs, err := m.ipfs.Refs(hash, true, true)
	if err != nil {
		return "", err
	}
	var (
		// will determine if we have updated the root hash
		rootUpdated bool
		// will determine if we have updated a ref hash
		refsUpdated bool
	)
	// iterate over all references for the requested hash
	// and mark whether or not a reference hasn't been seen before
	for _, ref := range refs {
		if !object.UploadedRefs[ref] {
			// if we haven't updated a ref yet, mark as true
			if !refsUpdated {
				refsUpdated = true
			}
			// mark this particular ref as being seen
			object.UploadedRefs[ref] = true
		}
	}
	// check if we need to cupdate the roots used
	if !object.UploadedRootNodes[hash] {
		// update customer object with root
		object.UploadedRootNodes[hash] = true
		// mark as having updated the root
		rootUpdated = true
	}
	// if we have no refs to calculate size for, it means this upload
	// will consume no additional storage space, thus we can avoid charging them
	// all but the datasize of the root (if at all)
	if !refsUpdated {
		// if we have updated the used roots of the customer object
		// this means we need to first update the customer object
		// stored in database
		if rootUpdated {
			resp, err := m.put(username, &object)
			if err != nil {
				return "", err
			}
			// return the new customer object hash and the datasize
			return resp, nil
		}
		// since we haven't updated the root, nor a reference
		// simply return
		return "", nil
	}
	// store new customer object in ipfs
	// update database, and return hash of new object
	resp, err := m.put(username, &object)
	if err != nil {
		return "", err
	}
	return resp, nil
}

// put is a wrapper for commonly used functionality.
// it is responsible for putting the object to ipfs, and updating
// the associated usermodel in database
func (m *Manager) put(username string, obj *Object) (string, error) {
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	resp, err := m.ipfs.DagPut(marshaled, "json", "cbor")
	if err != nil {
		return "", err
	}
	if err := m.ipfs.Pin(resp); err != nil {
		return "", err
	}
	if err := m.um.UpdateCustomerObjectHash(username, resp); err != nil {
		return "", err
	}
	return resp, nil
}
