package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type IPFSPrivateNetwork struct {
	gorm.Model
	Name     string         `gorm:"type:varchar(255)"`
	APIURL   string         `gorm:"type:varchar(255)"`
	SwarmKey string         `gorm:"type:varchar(255)"`
	Users    pq.StringArray `gorm:"type:text[]"` // these are the users to which this IPFS network connection applies to specified by eth address
	IsHosted bool           `gorm:"type:boolean"`
	Network  Network
}

type Network struct {
	gorm.Model
	LocalNodeAddresses pq.StringArray `gorm:"type:text[]"` // these are the nodes whichwe run, and can connect to
	LocalNodePeerIDs   pq.StringArray `gorm:"type:text[];column:local_node_peer_ids"`
	Hosted             Hosted
}

type Hosted struct {
	gorm.Model
	BootstrapPeers   pq.StringArray `gorm:"type:text[]"` // these are the boostrap peers which we run
	BootstrapPeerIDs pq.StringArray `gorm:"type:text[];column:bootstrap_peer_ids"`
}

type IPFSNetworkManager struct {
	DB *gorm.DB
}

func NewIPFSNetworkManager(db *gorm.DB) *IPFSNetworkManager {
	return &IPFSNetworkManager{DB: db}
}

// TODO: Add in multiformat address validation
func (im *IPFSNetworkManager) CreatePrivateNetwork(name, apiURL, swarmKey string, isHosted bool, arrayParameters map[string]*[]string, users []string) (*IPFSPrivateNetwork, error) {
	var pnet IPFSPrivateNetwork
	/*	if check := im.preload().Where("name = ?").First(&net); check.Error != nil {
		return nil, check.Error
	}*/
	if pnet.CreatedAt != nilTime {
		return nil, errors.New("network already exists")
	}
	pnet.Name = name
	pnet.APIURL = apiURL
	// if we were passed in a list of users, then we will add them
	if len(users) > 0 {
		for _, v := range users {
			pnet.Users = append(pnet.Users, v)
		}
	}
	// if were hosting the network infrastructure ourselves set that up too
	if isHosted {
		pnet.IsHosted = true
		bPeers := *arrayParameters["bootstrap_peers"]
		nodeAddresses := *arrayParameters["local_node_addresses"]
		if len(bPeers) != len(nodeAddresses) {
			return nil, errors.New("bootstrap peers and node ip address are not equal length")
		}
		for k, v := range bPeers {
			pnet.Network.Hosted.BootstrapPeers = append(pnet.Network.Hosted.BootstrapPeers, v)
			pnet.Network.LocalNodeAddresses = append(pnet.Network.LocalNodeAddresses, nodeAddresses[k])
			//pnet.Hosted.BootstrapPeers = append(net.Hosted.BootstrapPeers, v)
			//pnet.NetHosted.LocalNodeIPAddresses = append(net.Hosted.LocalNodeIPAddresses, nodeIPAddresses[k])
		}
	} else {
		nodeAddresses := arrayParameters["local_node_addresses"]
		for _, v := range *nodeAddresses {
			pnet.Network.LocalNodeAddresses = append(pnet.Network.LocalNodeAddresses, v)
		}
	}
	if check := im.DB.Create(&pnet); check.Error != nil {
		return nil, check.Error
	}
	return &pnet, nil
}

func (im *IPFSNetworkManager) preload() *gorm.DB {
	// we do this so we can preload the HostedNetwork relations
	return im.DB.Preload("Hosted")
}
