package dccd

/*

	DCCD is a utility used to request a particular content hash from all known IPFS gateways.
	The intended purpose is to spread content throughout the IPFS network cache.
	Additional features will be requesting content from user specified nodes as well, allowing for additional cache dispersion
*/

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
	"github.com/jinzhu/gorm"
)

type DCCDManager struct {
	Shell    *ipfsapi.Shell
	Gateways map[string]int
	DB       *gorm.DB
}

func NewDCCDManager(connectionURL string) *DCCDManager {
	if connectionURL == "" {
		// load a default api
		connectionURL = "localhost:5001"
	}
	return &DCCDManager{Shell: ipfsapi.NewShell(connectionURL)}
}

func (dc *DCCDManager) ConnecToDatabase(db *gorm.DB) {
	dc.DB = db
}

func (dc *DCCDManager) ParseGateways() {
	indexes := make(map[string]int)
	for k, v := range gateArrays {
		indexes[v] = k
	}
	dc.Gateways = indexes
}

func (dc *DCCDManager) DisperseContent(contentHash string) (map[string]bool, error) {
	dispersals := make(map[string]bool)

	if len(dc.Gateways) < 1 {
		return nil, errors.New("please parse gateways before dispersing content")
	}
	//var err error
	for k := range dc.Gateways {
		url := fmt.Sprintf("%s/%s", k, contentHash)
		resp, err := http.Get(url)
		if err != nil {
			dispersals[k] = false
			continue
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			dispersals[k] = false
			continue
		}
		if string(body) != "Hello from IPFS Gateway Checker" {
			dispersals[k] = false
		}
		dispersals[k] = true
	}
	return dispersals, nil
}
