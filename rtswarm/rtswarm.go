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

// NewSwarmManager is used to generate our swarm manager helper interface
func NewSwarmManager() (*SwarmManager, error) {
	sm := &SwarmManager{}
	sm.GenSwarmAPIConfig()
	if err := sm.GenSwarmPrivateKeys(); err != nil {
		return nil, err
	}
	sm.GenSwarmClient()
	return sm, nil
}

// GenSwarmAPIConfig is used to generate a default swarm api configuration
func (sm *SwarmManager) GenSwarmAPIConfig() {
	sm.Config = api.NewDefaultConfig()
}

// GenSwarmPrivateKeys is used to generate our swarm private keys
func (sm *SwarmManager) GenSwarmPrivateKeys() error {
	key, err := crypto.GenerateKey()
	if err != nil {
		return err
	}
	sm.PrivateKey = key
	return nil
}

// GenSwarmClient is used to generate our swarm client
func (sm *SwarmManager) GenSwarmClient() {
	sm.Client = client.NewClient(client.DefaultGateway)
}
