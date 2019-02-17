package customer

import (
	"encoding/json"

	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/rtfs"
)

// Manager is used to handle managing customer objects
type Manager struct {
	um   *models.UserManager
	ipfs rtfs.Manager
}

// Object represents a single customer
type Object struct {
	// UploadedRefs is a map used to check whether or not a particular link ref has been uploaded
	// the only time this value is populated is when an ipld objecct (say a folder) is uploaded, all
	// the links of the folder will fill this map
	UploadedRefs map[string]bool `json:"uploaded_nodes"`
	// UploadedRootNodes is a map used to check whether or not a particular entire root node has been uploaded.
	// each upload will always count as an uploaded root node however only ipld objects that contain links will also go to uploaded refs
	UploadedRootNodes map[string]bool `json:"uploaded_root_nodes"`
}

/* the overall process for deduplicated billing will look like:

file uploads:
1) add file to ipfs, record hash
2) look up references of the hash, record if any
3) check customer object root nodes for hash recorded in step 1
3a) if match with no references skip processing
3b) if match with references, check references
3c) if no match, with no references charge
3d) if mo match, with references, check references
4) process references (search through a users `UploadedRefs` for anything that is false, they get charged)
*/

// Unmarshal is used to take an object hash, and unmarshal it into a typed object
func (m *Manager) Unmarshal(hash string, out *Object) error {
	return m.ipfs.DagGet(hash, out)
}

// GetDeduplicatedStorageSpaceInBytes is used to get the deduplicated storage space used
// by this particular hash. It calculates deduplicated storage costs based on previous uploads
// the user has made, and does not consider upload from other users.
// it returns the new customer object hash of the user, and updates it.
func (m *Manager) GetDeduplicatedStorageSpaceInBytes(username, hash string) (string, int, error) {
	// find the user model
	user, err := m.um.FindByUserName(username)
	if err != nil {
		return "", 0, err
	}
	// construct an empty object
	var object Object
	// unmarshal the customer object hash stored in user model into a typed object
	if err := m.Unmarshal(user.CustomerObjectHash, &object); err != nil {
		return "", 0, err
	}
	// if the customer object is empty, then we return full storage space
	// consumed by the given hash
	if len(object.UploadedRefs) == 0 && len(object.UploadedRootNodes) == 0 {
		object = Object{
			UploadedRefs:      make(map[string]bool),
			UploadedRootNodes: make(map[string]bool),
		}
		// get the size
		size, _, err := rtfs.DedupAndCalculatePinSize(hash, m.ipfs)
		if err != nil {
			return "", 0, err
		}
		// get references for the requested hash
		refs, err := m.ipfs.Refs(hash, true, true)
		if err != nil {
			return "", 0, err
		}
		// iterate over all references marking them as being used
		for _, v := range refs {
			object.UploadedRefs[v] = true
		}
		// mark the root node as being used
		object.UploadedRootNodes[hash] = true
		// marshal object
		marshaled, err := json.Marshal(&object)
		if err != nil {
			return "", 0, err
		}
		// store object
		resp, err := m.ipfs.DagPut(marshaled, "json", "cbor")
		if err != nil {
			return "", 0, err
		}
		// pin the object so we have access to it long-term
		if err := m.ipfs.Pin(resp); err != nil {
			return "", 0, err
		}
		// update the customer object hash stored in database
		if err := m.um.UpdateCustomerObjectHash(username, resp); err != nil {
			return "", 0, err
		}
		return resp, int(size), nil
	}
	// search through the refs for the given hash
	refs, err := m.ipfs.Refs(hash, true, true)
	if err != nil {
		return "", 0, err
	}
	// refsToCalculate will hold all references that we need
	// to calculate th estorage size for
	var (
		refsToCalculate []string
		size            int
	)
	for _, ref := range refs {
		if !object.UploadedRefs[ref] {
			refsToCalculate = append(refsToCalculate, ref)
		}
	}
	// check if we need to calculate datasize of the roots
	if !object.UploadedRootNodes[hash] {
		stats, err := m.ipfs.Stat(hash)
		if err != nil {
			return "", 0, err
		}
		// update size
		size = stats.DataSize
		// update customer object with root
		object.UploadedRootNodes[hash] = true
	}
	// if we have no refs to calculate size for, it means this upload
	// will consume no additional storage space, thus we can avoid charging them
	// all but the datasize of the root (if at all)
	if len(refsToCalculate) == 0 {
		return user.CustomerObjectHash, size, nil
	}
	// calculate size of all references
	// also use this to update the refs of customer object
	for _, ref := range refsToCalculate {
		stats, err := m.ipfs.Stat(ref)
		if err != nil {
			return "", 0, err
		}
		size = size + stats.DataSize
		object.UploadedRefs[ref] = true
	}
	// marshal the updated customer object
	marshaled, err := json.Marshal(&object)
	if err != nil {
		return "", 0, err
	}
	// put the new object
	resp, err := m.ipfs.DagPut(marshaled, "json", "cbor")
	// pin the new object
	if err := m.ipfs.Pin(resp); err != nil {
		return "", 0, err
	}
	// update user model in database
	if err := m.um.UpdateCustomerObjectHash(username, resp); err != nil {
		return "", 0, err
	}
	return resp, size, nil
}

// UpdateObject is used to update the customer object for a user
func (m *Manager) UpdateObject(username string, rootNodes, refs []string) error {
	return nil
}
