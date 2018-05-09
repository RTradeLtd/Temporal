package rtfs_cluster

import (
	"github.com/ipfs/ipfs-cluster/api/rest/client"
	//"github.com/ipfs/ipfs-cluster/api/rest/client"
)

// ClusterManager is a helper interface to interact with the cluster apis
type ClusterManager struct {
	ClusterConfig *client.Config
	ClusterClient *client.Client
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
	cm.ClusterConfig = &client.Config{}
}

// GenClient is used to generate a client to interact with the cluster
func (cm *ClusterManager) GenClient() error {
	cl, err := client.NewClient(cm.ClusterConfig)
	if err != nil {
		return err
	}
	cm.ClusterClient = cl
	return nil
}

/*
func BuildCluster() {
	host, cfg := BuildClusterHost()
	ipfsc.NewCluster(host, cfg)
}*/
