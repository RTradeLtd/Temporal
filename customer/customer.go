package customer

import (
	"github.com/RTradeLtd/gorm"
	"github.com/RTradeLtd/rtfs"
)

// Manager is used to handle managing customer objects
type Manager struct {
	DB   *gorm.DB
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
	Username          string          `json:"user_name"`
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
