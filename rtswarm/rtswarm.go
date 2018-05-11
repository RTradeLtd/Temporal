package rtswarm

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/swarm/api"
	"github.com/ethereum/go-ethereum/swarm/api/client"
)

// SwarmManager is a helper interface
type SwarmManager struct {
	Config     *api.Config
	Client     *client.Client
	PrivateKey *ecdsa.PrivateKey
}

// Initialize is used to initialize the swarm manager
func Initialize() (*SwarmManager, error) {
	sm := &SwarmManager{}
	sm.GenDefaultConfig()
	sm.GenPrivateKey()
	return sm, nil
}

// GenDefaultConfig is used to generate default swarm config
func (sm SwarmManager) GenDefaultConfig() {
	sm.Config = api.NewDefaultConfig()
}

// GenPrivateKey is used to generate a private key pair
func (sm SwarmManager) GenPrivateKey() error {
	key, err := crypto.GenerateKey()
	if err != nil {
		return err
	}
	sm.PrivateKey = key
	return nil
}

// InitializeConfiguration is used to finalize the config build process
// TODO: Add info and research waht this all entails
func (sm SwarmManager) InitializeConfiguration() {
	sm.Config.Init(sm.PrivateKey)
}

// GenerateAPIClient is used to generate an api client for swarm
func (sm SwarmManager) GenerateAPIClient(gateway string) {

	if gateway == "" {
		sm.Client = client.NewClient(client.DefaultGateway)
	} else {
		sm.Client = client.NewClient(gateway)
	}
}
