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
	return bpeers, nil
}
