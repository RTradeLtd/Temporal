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
