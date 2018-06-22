package rtfsp

import (
	"encoding/json"
	"os"

	addrUtil "github.com/ipfs/go-ipfs-addr"
	cg "github.com/ipfs/go-ipfs/repo/config"
)

type PrivateConfigManager struct {
	Config *cg.Config
}

const (
	// ConfigFilePath is where the ipfs config file is located
	ConfigFilePath = "/ipfs/config"
	// SwarmKeyPath is the location of our swarm key
	SwarmKeyPath = "/ipfs/swarm.key"
)

// GenerateConfigManager generates the config manager for our private
func GenerateConfigManager(configFilePath string) (*PrivateConfigManager, error) {
	var conf cg.Config
	file, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(file).Decode(&conf)
	pcm := PrivateConfigManager{Config: &conf}
	return &pcm, err
}

func (pcm *PrivateConfigManager) GenerateIPFSMultiAddr(address string) (addrUtil.IPFSAddr, error) {
	ipfsAddr, err := addrUtil.ParseString(address)
	if err != nil {
		return nil, err
	}
	return ipfsAddr, nil
}

func (pcm *PrivateConfigManager) GenerateBootstrapPeer(address string) (cg.BootstrapPeer, error) {
	bpeer, err := cg.ParseBootstrapPeer(address)
	if err != nil {
		return nil, err
	}
	return bpeer, nil
}

// ConfigureBootstrap is a variadic function used to generate a list of bootstrap peers
// it will validate all provided peers, and fail if it detects an error
func (pcm *PrivateConfigManager) ConfigureBootstrap(peers ...string) ([]cg.BootstrapPeer, error) {
	var bpeers []cg.BootstrapPeer
	for _, v := range peers {
		bpeer, err := pcm.GenerateBootstrapPeer(v)
		if err != nil {
			return nil, err
		}
		bpeers = append(bpeers, bpeer)
	}
	pcm.Config.SetBootstrapPeers(bpeers)
	return bpeers, nil
}

// ParseConfigAndWrite is used to parse a configuration object and write it
func (pcm *PrivateConfigManager) ParseConfigAndWrite(peers ...string) error {
	// Create file, truncating if it exists
	cFilePath, err := os.Open(ConfigFilePath)
	if err != nil {
		return err
	}
	cfg, err := cg.Init(cFilePath, 4192)
	if err != nil {
		return err
	}
	bootPeers, err := pcm.ConfigureBootstrap(peers...)
	if err != nil {
		return err
	}
	cfg.SetBootstrapPeers(bootPeers)
	pcm.Config = cfg
	return nil
}
