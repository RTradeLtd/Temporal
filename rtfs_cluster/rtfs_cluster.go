package rtfs_cluster

import (
	"fmt"
	"log"

	gocid "github.com/ipfs/go-cid"
	"github.com/ipfs/ipfs-cluster/api"
	"github.com/ipfs/ipfs-cluster/api/rest/client"
	//"github.com/ipfs/ipfs-cluster/api/rest/client"
)

// ClusterManager is a helper interface to interact with the cluster apis
type ClusterManager struct {
	Config *client.Config
	Client *client.Client
}

// Initialize is used to init, and return a cluster manager object
func Initialize(hostAddress, hostPort string) (*ClusterManager, error) {
	cm := ClusterManager{}
	cm.GenRestAPIConfig()
	if hostAddress != "" && hostPort != "" {
		cm.Config.Host = hostAddress
		cm.Config.Port = hostPort
	}
	// modify default config with infrastructure specific settings
	err := cm.GenClient()
	if err != nil {
		return nil, err
	}
	return &cm, nil
}

// GenRestAPIConfig is used to generate the api cfg
// needed to interact with the cluster
func (cm *ClusterManager) GenRestAPIConfig() {
	cm.Config = &client.Config{}
}

// GenClient is used to generate a client to interact with the cluster
func (cm *ClusterManager) GenClient() error {
	cl, err := client.NewClient(cm.Config)
	if err != nil {
		return err
	}
	cm.Client = cl
	return nil
}

// ParseLocalStatusAllAndSync is used to parse through any errors
// and resync them
// TODO: make more robust
func (cm *ClusterManager) ParseLocalStatusAllAndSync() ([]*gocid.Cid, error) {
	// this will hold all the cids that have been synced
	var syncedCids []*gocid.Cid
	// only fetch the local status for all pins
	pinInfo, err := cm.Client.StatusAll(true)
	if err != nil {
		return nil, err
	}
	// parse through the pin status
	for _, v := range pinInfo {
		cid := v.Cid
		// fetch a mapping of all peers and their status (in this case only 1 will be present)
		peermap := v.PeerMap
		// get the client ID of the local IPFS Cluster node
		id, err := cm.Client.ID()
		if err != nil {
			return nil, err
		}
		// fetch the pin info for this node only
		globalPinInfo := peermap[id.ID]
		// get a list of the errors
		errString := globalPinInfo.Error
		// if there are none, then skip processing this cid
		if errString == "" {
			continue
		}
		// we have an error, so lets fix that
		_, err = cm.Client.Sync(cid, true)
		if err != nil {
			log.Fatal(err)
		}
		// add the cid to the list of synced ones
		syncedCids = append(syncedCids, cid)
	}
	return syncedCids, nil
}

// RemovePinFromCluster is used to remove a pin from the cluster
func (cm *ClusterManager) RemovePinFromCluster(cidString string) error {
	decoded, err := cm.DecodeHashString(cidString)
	if err != nil {
		return err
	}
	err = cm.Client.Unpin(decoded)
	if err != nil {
		return err
	}
	return nil
}

// FetchLocalStatus is used to fetch the local status of all pins
func (cm *ClusterManager) FetchLocalStatus() (map[*gocid.Cid]string, error) {
	var response = make(map[*gocid.Cid]string)
	pinInfo, err := cm.Client.StatusAll(true)
	if err != nil {
		return response, err
	}
	for _, v := range pinInfo {
		cid := v.Cid
		peermap := v.PeerMap
		id, err := cm.Client.ID()
		if err != nil {
			return response, err
		}
		globalPinInfo := peermap[id.ID]
		errString := globalPinInfo.Error
		response[cid] = errString
	}
	return response, nil
}

// GetStatusForCidLocally is used to fetch the local status for a particular cid
func (cm *ClusterManager) GetStatusForCidLocally(cidString string) (*api.GlobalPinInfo, error) {
	decoded, err := cm.DecodeHashString(cidString)
	if err != nil {
		return nil, err
	}
	status, err := cm.Client.Status(decoded, true)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// GetStatusForCidGlobally is used to fetch the global status for a particular cid
func (cm *ClusterManager) GetStatusForCidGlobally(cidString string) (*api.GlobalPinInfo, error) {
	decoded, err := cm.DecodeHashString(cidString)
	if err != nil {
		return nil, err
	}
	status, err := cm.Client.Status(decoded, false)
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// ListPeers is used to list the known cluster peers
func (cm *ClusterManager) ListPeers() ([]api.ID, error) {
	peers, err := cm.Client.Peers()
	if err != nil {
		return nil, err
	}
	return peers, nil
}

// DecodeHashString is used to take a hash string, and turn it into a CID
func (cm *ClusterManager) DecodeHashString(cidString string) (*gocid.Cid, error) {
	cid, err := gocid.Decode(cidString)
	if err != nil {
		return nil, err
	}
	return cid, nil
}

// Pin is used to add a pin to the cluster
func (cm *ClusterManager) Pin(cid *gocid.Cid) error {
	err := cm.Client.Pin(cid, -1, -1, cid.String())
	if err != nil {
		return err
	}
	status, err := cm.Client.Status(cid, true)
	if err != nil {
		fmt.Println("error pinning hash to cluster")
		return err
	}
	fmt.Println("status")
	fmt.Println(status)
	return nil
}
