package rtfs_cluster

import (
	"fmt"
	"log"

	gocid "github.com/ipfs/go-cid"
	"github.com/ipfs/ipfs-cluster/api/rest/client"
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
func (cm *ClusterManager) ParseLocalStatusAllAndSync() error {
	pinInfo, err := cm.Client.StatusAll(true)
	if err != nil {
		return err
	}
	for _, v := range pinInfo {
		cid := v.Cid
		_, err := cm.Client.Sync(cid, true)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
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
