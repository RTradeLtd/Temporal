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
	Hosted   HostedNetwork
}

type HostedNetwork struct {
	gorm.Model
	BootstrapPeers       pq.StringArray `gorm:"type:text[]"` // these are the boostrap peers which we run
	LocalNodeIPAddresses pq.StringArray `gorm:"type:text[]"` // these are the nodes whichwe run, and can connect to
}

type IPFSNetworkManager struct {
	DB *gorm.DB
}

func NewIPFSNetworkManager(db *gorm.DB) *IPFSNetworkManager {
	return &IPFSNetworkManager{DB: db}
}

func (im *IPFSNetworkManager) CreatePrivateNetwork(name, apiURL, swarmKey string, isHosted bool, arrayParameters map[string]interface{}, users ...string) (*IPFSPrivateNetwork, error) {
	var net IPFSPrivateNetwork
	/*	if check := im.preload().Where("name = ?").First(&net); check.Error != nil {
		return nil, check.Error
	}*/
	if net.CreatedAt != nilTime {
		return nil, errors.New("network already exists")
	}
	net.Name = name
	net.APIURL = apiURL
	// if we were passed in a list of users, then we will add them
	if len(users) > 0 {
		for _, v := range users {
			net.Users = append(net.Users, v)
		}
	}
	// if were hosting the network infrastructure ourselves set that up too
	if isHosted {
		net.IsHosted = true
		bPeers, ok := arrayParameters["bootstrap_peers"].([]string)
		if !ok {
			return nil, errors.New("bootstrap_peers is not of type []string")
		}
		nodeIPAddresses, ok := arrayParameters["local_node_ip_addresses"].([]string)
		if !ok {
			return nil, errors.New("local_node_ip_address is not of type []string")
		}
		if len(bPeers) != len(nodeIPAddresses) {
			return nil, errors.New("bootstrap peers and node ip address are not equal length")
		}
		for k, v := range bPeers {
			net.Hosted.BootstrapPeers = append(net.Hosted.BootstrapPeers, v)
			net.Hosted.LocalNodeIPAddresses = append(net.Hosted.LocalNodeIPAddresses, nodeIPAddresses[k])
		}
	}
	im.DB.Create(&net)
	return &net, nil
}

func (im *IPFSNetworkManager) preload() *gorm.DB {
	// we do this so we can preload the HostedNetwork relations
	return im.DB.Preload("Hosted")
}
