package rtfs_cluster

import (
	"fmt"
	"log"

	gocid "github.com/ipfs/go-cid"
	"github.com/ipfs/ipfs-cluster/api"
	"github.com/ipfs/ipfs-cluster/api/rest/client"
	ma "github.com/multiformats/go-multiaddr"
	//"github.com/ipfs/ipfs-cluster/api/rest/client"
)

// ClusterManager is a helper interface to interact with the cluster apis
type ClusterManager struct {
	Config *client.Config
	Client *client.Client
}

// Initialize is used to init, and return a cluster manager object
func Initialize() *ClusterManager {
	cm := ClusterManager{}
	cm.GenRestAPIConfig()
	// modify default config with infrastructure specific settings
	cm.GenClient()
	return &cm
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

// ParseStatusAllSyncErrors is used to parse through any errors
// and resync them
func (cm *ClusterManager) ParseLocalStatusAllAndSync() ([]*gocid.Cid, error) {
	var syncedCids []*gocid.Cid
	pinInfo, err := cm.Client.StatusAll(true)
	if err != nil {
		return nil, err
	}
	for _, v := range pinInfo {
		cid := v.Cid
		peermap := v.PeerMap
		id, err := cm.Client.ID()
		if err != nil {
			return nil, err
		}
		globalPinInfo := peermap[id.ID]
		errString := globalPinInfo.Error
		fmt.Println(globalPinInfo)
		fmt.Println(errString)
		if errString == "" {
			continue
		}
		_, err = cm.Client.Sync(cid, true)
		if err != nil {
			log.Fatal(err)
		}
		syncedCids = append(syncedCids, cid)
	}
	return syncedCids, nil
}

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
		if errString == "" {
			continue
		}
		response[cid] = errString
	}
	return response, nil
}
func (cm *ClusterManager) GetStatusForCidLocally(cidString string) (*api.GlobalPinInfo, error) {
	decoded := cm.DecodeHashString(cidString)
	status, err := cm.Client.Status(decoded, true)
	if err != nil {
		return nil, err
	}
	fmt.Println(status)
	return &status, nil
}

func (cm *ClusterManager) GetStatusForCidGlobally(cidString string) (*api.GlobalPinInfo, error) {
	decoded := cm.DecodeHashString(cidString)
	status, err := cm.Client.Status(decoded, false)
	if err != nil {
		return nil, err
	}
	fmt.Println(status)
	return &status, nil
}

func (cm *ClusterManager) ListPeers() ([]api.ID, error) {
	peers, err := cm.Client.Peers()
	if err != nil {
		return nil, err
	}
	return peers, nil
}

// AddPeerToCluster is used to add a peer to the cluster
// TODO: still needs to be completed
func (cm *ClusterManager) AddPeerToCluster(addr ma.Multiaddr) {
	cm.Client.PeerAdd(addr)
}

func (cm *ClusterManager) DecodeHashString(cidString string) *gocid.Cid {
	cid, err := gocid.Decode(cidString)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cid)
	return cid
}

func (cm *ClusterManager) Pin(cid *gocid.Cid) error {
	err := cm.Client.Pin(cid, -1, -1, "test")
	if err != nil {
		return err
	}
	stat, err := cm.Client.Status(cid, true)
	if err != nil {
		return err
	}
	fmt.Println(stat)
	fmt.Printf("%+v\n", stat)
	return nil
}

/*
func BuildCluster() {
	host, cfg := BuildClusterHost()
	ipfsc.NewCluster(host, cfg)
}*/
