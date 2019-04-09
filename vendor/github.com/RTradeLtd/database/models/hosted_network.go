package models

import (
	"errors"
	"fmt"
	"time"

	"github.com/RTradeLtd/database/utils"
	"github.com/RTradeLtd/gorm"
	"github.com/lib/pq"
)

// HostedNetwork is a private network for which we are responsible of the infrastructure
type HostedNetwork struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Name      string     `gorm:"unique;type:varchar(255)"` // Name of the network node
	Activated *time.Time // Activated represents the most recent activation, null if offline
	Disabled  bool

	PeerKey string // Private key used to generate peerID for this network node

	// SwarmAddr is the address of swarm port. Slated for deprecation if HTTP path
	// support is added to the multiaddr spec and go-multiaddr
	SwarmAddr string `gorm:"type:varchar(255)"`
	// SwarmKey is the key used to connect to this peer
	SwarmKey string `gorm:"type:varchar(255)"`

	// Used to set Allowed-Origin headers on API requests
	APIAllowedOrigin string `gorm:"type:varchar(255)"`

	// Toggles whether gateway should be exposed through Nexus delegator
	GatewayPublic bool `gorm:"type:boolean"`

	// Peers to bootstrap node onto
	BootstrapPeerAddresses pq.StringArray `gorm:"type:text[]"`
	BootstrapPeerIDs       pq.StringArray `gorm:"type:text[];column:bootstrap_peer_ids"`

	// Resources for deployed node
	ResourcesCPUs     int
	ResourcesDiskGB   int
	ResourcesMemoryGB int

	// Owner is the creator of the private network, and is allowed to invoke
	// administrative commands, such as network destruction.
	Owners pq.StringArray `gorm:"type:text[]"`
	// Users allowed to control this node. Includes API access.
	Users pq.StringArray `gorm:"type:text[]"` // these are the users to which this IPFS network connection applies to specified by eth address
}

// HostedNetworkManager is used to manipulate IPFS network models in the database
type HostedNetworkManager struct {
	DB *gorm.DB
}

// NewHostedNetworkManager is used to initialize our database connection
func NewHostedNetworkManager(db *gorm.DB) *HostedNetworkManager {
	return &HostedNetworkManager{DB: db}
}

// GetNetworkByName is used to retrieve a network from the database based off of its name
func (im *HostedNetworkManager) GetNetworkByName(name string) (*HostedNetwork, error) {
	var pnet HostedNetwork
	if check := im.DB.Model(&pnet).Where("name = ?", name).First(&pnet); check.Error != nil {
		return nil, check.Error
	}
	return &pnet, nil
}

// SwarmDetails provides data about IPFS swarm connection
type SwarmDetails struct {
	Addr string
	Key  string
}

// GetSwarmDetails is used to retrieve data about IPFS swarm connection
func (im *HostedNetworkManager) GetSwarmDetails(network string) (*SwarmDetails, error) {
	pnet, err := im.GetNetworkByName(network)
	if err != nil {
		return nil, err
	}
	return &SwarmDetails{
		Addr: pnet.SwarmAddr,
		Key:  pnet.SwarmKey,
	}, nil
}

// APIDetails provides data about IPFS API connection
type APIDetails struct {
	AllowedOrigin string
}

// GetAPIDetails is used to retrieve data about IPFS API connection
func (im *HostedNetworkManager) GetAPIDetails(network string) (*APIDetails, error) {
	pnet, err := im.GetNetworkByName(network)
	if err != nil {
		return nil, err
	}
	return &APIDetails{
		AllowedOrigin: pnet.APIAllowedOrigin,
	}, nil
}

// UpdateNetworkByName updates the given network with given attributes
func (im *HostedNetworkManager) UpdateNetworkByName(name string, attrs map[string]interface{}) error {
	var pnet HostedNetwork
	if check := im.DB.Model(&pnet).Where("name = ?", name).First(&pnet).Update(attrs); check.Error != nil {
		return check.Error
	}
	return nil
}

// SaveNetwork saves the given HostedNetwork in the database
func (im *HostedNetworkManager) SaveNetwork(n *HostedNetwork) error {
	if check := im.DB.Save(n); check != nil && check.Error != nil {
		return check.Error
	}
	return nil
}

// GetOfflineNetworks returns all currently offline networks
func (im *HostedNetworkManager) GetOfflineNetworks(disabled bool) ([]*HostedNetwork, error) {
	var networks = []*HostedNetwork{}
	var check = im.DB.Model(&HostedNetwork{}).
		Where("activated is null").
		Where("disabled = ?", disabled).
		Find(&networks)
	return networks, check.Error
}

// NetworkAccessOptions configures access to a hosted private network
type NetworkAccessOptions struct {
	Owner            string
	Users            []string
	APIAllowedOrigin string
	PublicGateway    bool
}

// CreateHostedPrivateNetwork is used to store a new hosted private network in the database
func (im *HostedNetworkManager) CreateHostedPrivateNetwork(
	name, swarmKey string,
	peers []string,
	access NetworkAccessOptions,
) (*HostedNetwork, error) {
	// check if network exists
	pnet := &HostedNetwork{}
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
	if access.Users != nil && len(access.Users) > 0 {
		for _, v := range access.Users {
			pnet.Users = append(pnet.Users, v)
		}
	}

	// set the owner of the network
	pnet.Owners = []string{access.Owner}

	// assign misc details
	pnet.Name = name
	pnet.SwarmKey = swarmKey
	pnet.APIAllowedOrigin = access.APIAllowedOrigin
	pnet.GatewayPublic = access.PublicGateway

	// create network entry
	if check := im.DB.Create(pnet); check.Error != nil {
		return nil, check.Error
	}
	return pnet, nil
}

// Delete is used to remove a network from the database
func (im *HostedNetworkManager) Delete(name string) error {
	net, err := im.GetNetworkByName(name)
	if err != nil {
		return err
	}
	return im.DB.Unscoped().Delete(net).Error
}
