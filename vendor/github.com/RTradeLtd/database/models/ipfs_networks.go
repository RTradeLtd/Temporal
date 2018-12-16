package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/RTradeLtd/database/utils"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
)

// HostedIPFSPrivateNetwork is a private network for which we are responsible of the infrastructure
type HostedIPFSPrivateNetwork struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name      string         `gorm:"unique;type:varchar(255)"`
	APIURL    string         `gorm:"type:varchar(255)"`
	SwarmKey  string         `gorm:"type:varchar(255)"`
	Users     pq.StringArray `gorm:"type:text[]"` // these are the users to which this IPFS network connection applies to specified by eth address
	Activated time.Time

	BootstrapPeerAddresses pq.StringArray `gorm:"type:text[]"`
	BootstrapPeerIDs       pq.StringArray `gorm:"type:text[];column:bootstrap_peer_ids"`

	ResourcesCPUs     int
	ResourcesDiskGB   int
	ResourcesMemoryGB int

	// note: local addresses currently unused
	LocalNodePeerAddresses pq.StringArray `gorm:"type:text[]"` // these are the nodes whichwe run, and can connect to
	LocalNodePeerIDs       pq.StringArray `gorm:"type:text[];column:local_node_peer_ids"`
}

// IPFSNetworkManager is used to manipulate IPFS network models in the database
type IPFSNetworkManager struct {
	DB *gorm.DB
}

// NewHostedIPFSNetworkManager is used to initialize our database connection
func NewHostedIPFSNetworkManager(db *gorm.DB) *IPFSNetworkManager {
	return &IPFSNetworkManager{DB: db}
}

// GetNetworkByName is used to retrieve a network from the database based off of its name
func (im *IPFSNetworkManager) GetNetworkByName(name string) (*HostedIPFSPrivateNetwork, error) {
	var pnet HostedIPFSPrivateNetwork
	if check := im.DB.Model(&pnet).Where("name = ?", name).First(&pnet); check.Error != nil {
		return nil, check.Error
	}
	return &pnet, nil
}

// GetAPIURLByName is used to retrieve the API url for a private network by its network name
func (im *IPFSNetworkManager) GetAPIURLByName(name string) (string, error) {
	pnet, err := im.GetNetworkByName(name)
	if err != nil {
		return "", err
	}
	return pnet.APIURL, nil
}

// UpdateNetworkByName updates the given network with given attributes
func (im *IPFSNetworkManager) UpdateNetworkByName(name string, attrs map[string]interface{}) error {
	var pnet HostedIPFSPrivateNetwork
	if check := im.DB.Model(&pnet).Where("name = ?", name).First(&pnet).Update(attrs); check.Error != nil {
		return check.Error
	}
	return nil
}

// CreateHostedPrivateNetwork is used to store a new hosted private network in the database
func (im *IPFSNetworkManager) CreateHostedPrivateNetwork(name, swarmKey string, peers, users []string) (*HostedIPFSPrivateNetwork, error) {
	// check if network exists
	pnet := &HostedIPFSPrivateNetwork{}
	if check := im.DB.Where("name = ?", name).First(pnet); check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return nil, check.Error
	}
	if pnet.CreatedAt != nilTime {
		return nil, errors.New("private network already exists")
	}

	// parse peers
	if peers != nil {
		for _, v := range peers {
			// parse peer address
			addr, err := utils.GenerateMultiAddrFromString(v)
			if err != nil {
				return nil, err
			}
			valid, err := utils.ParseMultiAddrForIPFSPeer(addr)
			if err != nil {
				return nil, err
			}
			if !valid {
				return nil, fmt.Errorf("provided peer '%s' is not a valid bootstrap peer", addr)
			}

			// parse peer ID
			peer := addr.String()
			formattedBAddr, err := utils.GenerateMultiAddrFromString(peer)
			if err != nil {
				return nil, err
			}
			parsedBPeerID, err := utils.ParsePeerIDFromIPFSMultiAddr(formattedBAddr)
			if err != nil {
				return nil, err
			}

			// register peer
			pnet.BootstrapPeerAddresses = append(pnet.BootstrapPeerAddresses, peer)
			pnet.BootstrapPeerIDs = append(pnet.BootstrapPeerIDs, parsedBPeerID)
		}
	}

	// parse authorized users
	if users != nil && len(users) > 0 {
		for _, v := range users {
			pnet.Users = append(pnet.Users, v)
		}
	} else {
		pnet.Users = append(pnet.Users, AdminAddress)
	}

	// assign name, swarm key and create network entry
	pnet.Name = name
	pnet.SwarmKey = swarmKey
	if check := im.DB.Create(pnet); check.Error != nil {
		return nil, check.Error
	}
	return pnet, nil
}

// Delete is used to remove a network from the database
func (im *IPFSNetworkManager) Delete(name string) error {
	net, err := im.GetNetworkByName(name)
	if err != nil {
		return err
	}
	return im.DB.Unscoped().Delete(net).Error
}
