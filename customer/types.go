package customer

import (
	"sync"

	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/rtfs"
)

// Manager is used to handle managing customer objects
type Manager struct {
	um    *models.UserManager
	ipfs  rtfs.Manager
	mutex sync.Mutex
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
