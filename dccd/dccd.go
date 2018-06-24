package dccd

/*

	DCCD is a utility used to request a particular content hash from all known IPFS gateways.
	The intended purpose is to spread content throughout the IPFS network cache.
	Additional features will be requesting content from user specified nodes as well, allowing for additional cache dispersion
*/

import (
	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
)

// This is a URL from which we can obtain a list of public gateways, taken from ipfg.sh
var PublicGatewayList = "https://raw.githubusercontent.com/ipfs/public-gateway-checker/master/gateways.json"

type DCCDManager struct {
	Shell *ipfsapi.Shell
}

func NewDCCDManager(connectionURL string) *DCCDManager {
	if connectionURL == "" {
		// load a default api
		connectionURL = "localhost:5001"
	}
}
