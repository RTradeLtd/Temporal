package rtfs

import ipfsapi "github.com/RTradeLtd/go-ipfs-api"

// newShell is used to establish our api shell for the ipfs node
func newShell(url string, direct bool) (sh *ipfsapi.Shell) {
	if direct {
		sh = ipfsapi.NewDirectShell(url)
	} else {
		if url == "" {
			sh = ipfsapi.NewShell("localhost:5001")
		} else {
			sh = ipfsapi.NewShell(url)
		}
	}
	return
}
