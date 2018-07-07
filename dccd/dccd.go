package dccd

/*

	DCCD is a utility used to request a particular content hash from all known IPFS gateways.
	The intended purpose is to spread content throughout the IPFS network cache.
	Additional features will be requesting content from user specified nodes as well, allowing for additional cache dispersion

	The initial idea is taken from `ipfg.sh` whose developer is Joss Brown (pseud.): https://github.com/JayBrown
	Source code for that script is in the `scripts` folder
*/

import (
	"errors"
	"fmt"

	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
)

type DCCDManager struct {
	Shell    *ipfsapi.Shell
	Gateways map[int]string
}

// NewDCCDManager establishes our initial connection to our local IPFS node
func NewDCCDManager(connectionURL string) *DCCDManager {
	if connectionURL == "" {
		// load a default api
		connectionURL = "localhost:5001"
	}
	shell := ipfsapi.NewShell(connectionURL)
	manager := &DCCDManager{Shell: shell}
	manager.ParseGateways()
	return manager
}

func (dc *DCCDManager) ParseGateways() {
	indexes := make(map[int]string)
	for k, v := range GateArrays {
		indexes[k] = v
	}
	dc.Gateways = indexes
}

func (dc *DCCDManager) ReconnectShell(connectionURL string) error {
	if connectionURL == "" {
		return errors.New("please provide a valid connection url")
	}
	shell := ipfsapi.NewShell(connectionURL)
	dc.Shell = shell
	return nil
}

func (dc *DCCDManager) DisperseContentWithShell(contentHash string) (map[string]bool, error) {
	m := make(map[string]bool)
	for _, v := range GateArrays {
		err := dc.ReconnectShell(v)
		r, err := dc.Shell.CatGet(contentHash)
		if err != nil {
			m[v] = false
			fmt.Println("dispersal failed for host ", v)
			continue
		}
		err = r.Close()
		if err != nil {
			fmt.Println("failed to close handler ", err)
		}
		fmt.Println("dispersal suceeded for host ", v)
		m[v] = true
	}
	return m, nil
}
