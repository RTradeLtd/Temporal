package models

import (
	"errors"

	"github.com/RTradeLtd/Temporal/utils"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

type HostedIPFSPrivateNetwork struct {
	gorm.Model
	Name                   string         `gorm:"type:varchar(255)"`
	APIURL                 string         `gorm:"type:varchar(255)"`
	SwarmKey               string         `gorm:"type:varchar(255)"`
	Users                  pq.StringArray `gorm:"type:text[]"` // these are the users to which this IPFS network connection applies to specified by eth address
	LocalNodePeerAddresses pq.StringArray `gorm:"type:text[]"` // these are the nodes whichwe run, and can connect to
	LocalNodePeerIDs       pq.StringArray `gorm:"type:text[];column:local_node_peer_ids"`
	BootstrapPeerAddresses pq.StringArray `gorm:"type:text[]"`
	BootstrapPeerIDs       pq.StringArray `gorm:"type:text[];column:bootstrap_peer_ids"`
}

type IPFSNetworkManager struct {
	DB *gorm.DB
}

func NewHostedIPFSNetworkManager(db *gorm.DB) *IPFSNetworkManager {
	return &IPFSNetworkManager{DB: db}
}

func (im *IPFSNetworkManager) GetNetworkByName(name string) (*HostedIPFSPrivateNetwork, error) {
	var pnet HostedIPFSPrivateNetwork
	if check := im.DB.Model(&pnet).Where("name = ?", name).First(&pnet); check.Error != nil {
		return nil, check.Error
	}
	return &pnet, nil
}

func (im *IPFSNetworkManager) GetAPIURLByName(name string) (string, error) {
	pnet, err := im.GetNetworkByName(name)
	if err != nil {
		return "", err
	}
	return pnet.APIURL, nil
}

// TODO: Validate swarm key and API url
func (im *IPFSNetworkManager) CreateHostedPrivateNetwork(name, apiURL, swarmKey string, arrayParameters map[string][]string, users []string) (*HostedIPFSPrivateNetwork, error) {
	pnet := &HostedIPFSPrivateNetwork{}
	if check := im.DB.Where("name = ?", name).First(pnet); check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return nil, check.Error
	}

	if pnet.CreatedAt != nilTime {
		return nil, errors.New("private network already exists")
	}

	bPeers := arrayParameters["bootstrap_peer_addresses"]
	nodeAddresses := arrayParameters["local_node_peer_addresses"]
	if len(bPeers) != len(nodeAddresses) {
		return nil, errors.New("bootstrap_peer_address and local_node_address length not equal")
	}
	for k, v := range bPeers {
		pnet.BootstrapPeerAddresses = append(pnet.BootstrapPeerAddresses, v)
		formattedBAddr, err := utils.GenerateMultiAddrFromString(v)
		if err != nil {
			return nil, err
		}
		parsedBPeerID, err := utils.ParsePeerIDFromIPFSMultiAddr(formattedBAddr)
		if err != nil {
			return nil, err
		}
		pnet.BootstrapPeerIDs = append(pnet.BootstrapPeerIDs, parsedBPeerID)
		pnet.LocalNodePeerAddresses = append(pnet.LocalNodePeerAddresses, nodeAddresses[k])
		formattedNAddr, err := utils.GenerateMultiAddrFromString(nodeAddresses[k])
		if err != nil {
			return nil, err
		}
		parsedNPeerID, err := utils.ParsePeerIDFromIPFSMultiAddr(formattedNAddr)
		if err != nil {
			return nil, err
		}
		pnet.LocalNodePeerIDs = append(pnet.LocalNodePeerIDs, parsedNPeerID)
	}
	if len(users) > 0 {
		for _, v := range users {
			pnet.Users = append(pnet.Users, v)
		}
	} else {
		pnet.Users = append(pnet.Users, AdminAddress)
	}

	pnet.Name = name
	pnet.APIURL = apiURL
	pnet.SwarmKey = swarmKey
	if check := im.DB.Create(pnet); check.Error != nil {
		return nil, check.Error
	}
	return pnet, nil
}
