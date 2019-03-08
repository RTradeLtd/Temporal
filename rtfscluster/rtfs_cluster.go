package rtfscluster

import (
	"context"
	"fmt"

	gocid "github.com/ipfs/go-cid"
	"github.com/ipfs/ipfs-cluster/api"
	"github.com/ipfs/ipfs-cluster/api/rest/client"
)

// ClusterManager is a helper interface to interact with the cluster apis
type ClusterManager struct {
	Config *client.Config
	Client client.Client
}

// Initialize is used to init, and return a cluster manager object
func Initialize(ctx context.Context, hostAddress, hostPort string) (*ClusterManager, error) {
	cm := ClusterManager{}
	cm.GenRestAPIConfig()
	if hostAddress != "" && hostPort != "" {
		cm.Config.Host = hostAddress
		cm.Config.Port = hostPort
	}
	// modify default config with infrastructure specific settings
	if err := cm.GenClient(); err != nil {
		return nil, err
	}
	if _, err := cm.ListPeers(ctx); err != nil {
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
	cl, err := client.NewDefaultClient(cm.Config)
	if err != nil {
		return err
	}
	cm.Client = cl
	return nil
}

// ListPeers is used to list the known cluster peers
func (cm *ClusterManager) ListPeers(ctx context.Context) ([]*api.ID, error) {
	peers, err := cm.Client.Peers(ctx)
	if err != nil {
		return nil, err
	}
	return peers, nil
}

// DecodeHashString is used to take a hash string, and turn it into a CID
func (cm *ClusterManager) DecodeHashString(cidString string) (gocid.Cid, error) {
	cid, err := gocid.Decode(cidString)
	if err != nil {
		return gocid.Cid{}, err
	}
	return cid, nil
}

// Pin is used to add a pin to the cluster
func (cm *ClusterManager) Pin(ctx context.Context, cid gocid.Cid) error {
	err := cm.Client.Pin(ctx, cid, api.PinOptions{ReplicationFactorMax: -1, ReplicationFactorMin: -1})
	if err != nil {
		return err
	}
	status, err := cm.Client.Status(ctx, cid, true)
	if err != nil {
		fmt.Println("error pinning hash to cluster")
		return err
	}
	fmt.Println("status")
	fmt.Println(status)
	return nil
}
