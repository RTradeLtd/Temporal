package models

import (
	"errors"
	"fmt"

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
	Network  Networks       `gorm:"ForeignKey:Name"`
}

type Networks struct {
	gorm.Model
	Name               string         `gorm:"type:varchar(255)"`
	LocalNodeAddresses pq.StringArray `gorm:"type:text[]"` // these are the nodes whichwe run, and can connect to
	LocalNodePeerIDs   pq.StringArray `gorm:"type:text[];column:local_node_peer_ids"`
	HostedNetwork      HostedNetworks `gorm:"ForeignKey:Name"`
}

type HostedNetworks struct {
	gorm.Model
	Name             string         `gorm:"type:varchar(255)"`
	BootstrapPeers   pq.StringArray `gorm:"type:text[]"` // these are the boostrap peers which we run
	BootstrapPeerIDs pq.StringArray `gorm:"type:text[];column:bootstrap_peer_ids"`
}

type IPFSNetworkManager struct {
	DB *gorm.DB
}

func NewIPFSNetworkManager(db *gorm.DB) *IPFSNetworkManager {
	return &IPFSNetworkManager{DB: db}
}

func (im *IPFSNetworkManager) GetNetworkByName(name string) (*IPFSPrivateNetwork, error) {
	var pnet IPFSPrivateNetwork
	if check := im.DB.Model(&pnet).Where("name = ?", name).First(&pnet).Model(&pnet.Network).
		Where("name = ?", name).First(&pnet.Network).Model(&pnet.Network.HostedNetwork).
		Where("name = ?", name).First(&pnet.Network.HostedNetwork); check.Error != nil {
		return nil, check.Error
	}
	return &pnet, nil
}

// TODO: Add in multiformat address validation
func (im *IPFSNetworkManager) CreatePrivateNetwork(name, apiURL, swarmKey string, isHosted bool, arrayParameters map[string][]string, users []string) (*IPFSPrivateNetwork, error) {
	var pnet IPFSPrivateNetwork
	if check := im.DB.Where("name = ?", name).First(&pnet); check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return nil, check.Error
	}
	if pnet.CreatedAt != nilTime {
		return nil, errors.New("network already exists")
	}
	im.DB.Model(&pnet).Association("Network")
	im.DB.Model(&pnet.Network).Association("HostedNetwork")
	pnet.Name = name
	pnet.APIURL = apiURL
	// if we were passed in a list of users, then we will add them
	if len(users) > 0 {
		for _, v := range users {
			pnet.Users = append(pnet.Users, v)
		}
	}
	hostedName := fmt.Sprintf("hosted+%s", name)
	privateName := fmt.Sprintf("private+%s", name)
	// if were hosting the network infrastructure ourselves set that up too
	if isHosted {
		pnet.IsHosted = true
		bPeers := arrayParameters["bootstrap_peer_addresses"]
		nodeAddresses := arrayParameters["local_node_addresses"]
		if len(bPeers) != len(nodeAddresses) {
			return nil, errors.New("bootstrap peers and node ip address are not equal length")
		}
		for k, v := range bPeers {
			pnet.Network.HostedNetwork.BootstrapPeers = append(pnet.Network.HostedNetwork.BootstrapPeers, v)
			pnet.Network.LocalNodeAddresses = append(pnet.Network.LocalNodeAddresses, nodeAddresses[k])
			//pnet.Hosted.BootstrapPeers = append(net.Hosted.BootstrapPeers, v)
			//pnet.NetHosted.LocalNodeIPAddresses = append(net.Hosted.LocalNodeIPAddresses, nodeIPAddresses[k])
		}
		pnet.Network.HostedNetwork.Name = privateName
	} else {
		nodeAddresses := arrayParameters["local_node_addresses"]
		for _, v := range nodeAddresses {
			pnet.Network.LocalNodeAddresses = append(pnet.Network.LocalNodeAddresses, v)
		}
	}
	pnet.Network.Name = hostedName
	/*var mn Networks
	var hn HostedNetworks
	mn.Name = name
	hn.Name = name
	fmt.Println(name)
	// we need to create the data in the database in order save, and link the mdoels
	if check := im.DB.Save(&mn); check.Error != nil {
		return nil, check.Error
	}
	if check := im.DB.Save(&hn); check.Error != nil {
		return nil, check.Error
	}*/
	// Relate the Network field of the model, to the networks table
	im.DB.Model(&pnet).Related(&pnet.Network, "Network")
	// Relate the hosted network field of the model to the hosted networks table
	im.DB.Model(&pnet.Network).Related(&pnet.Network.HostedNetwork, "HostedNetwork")
	if check := im.DB.Create(&pnet); check.Error != nil {
		return nil, check.Error
	}
	return &pnet, nil
}

func (im *IPFSNetworkManager) preload() *gorm.DB {
	// we do this so we can preload the HostedNetwork relations
	return im.DB.Preload("Network")
}
