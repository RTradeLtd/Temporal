package dccd

/*

	DCCD is a utility used to request a particular content hash from all known IPFS gateways.
	The intended purpose is to spread content throughout the IPFS network cache.
	Additional features will be requesting content from user specified nodes as well, allowing for additional cache dispersion
*/

import (
	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
)

type DCCDManager struct {
	Shell    *ipfsapi.Shell
	Gateways map[string]int
}

func NewDCCDManager(connectionURL string) *DCCDManager {
	if connectionURL == "" {
		// load a default api
		connectionURL = "localhost:5001"
	}
	return &DCCDManager{Shell: ipfsapi.NewShell(connectionURL)}
}

func (dc *DCCDManager) ParseGateways() {
	indexes := make(map[string]int)
	for k, v := range gateArrays {
		indexes[v] = k
	}
	dc.Gateways = indexes
}
