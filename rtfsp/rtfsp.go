package rtfsp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	addrUtil "github.com/ipfs/go-ipfs-addr"
	cg "github.com/ipfs/go-ipfs-config"
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
	return bpeers, nil
}

// ParseConfigAndWrite is used to parse a configuration object and write it
func (pcm *PrivateConfigManager) ParseConfigAndWrite(peers ...string) error {
	// Create file, truncating if it exists
	cFilePath, err := ioutil.TempFile("/tmp/", "tconfigfileforips")
	if err != nil {
		return err
	}
	cfg, err := cg.Init(cFilePath, 2048)
	if err != nil {
		return err
	}
	bootPeers, err := pcm.ConfigureBootstrap(peers...)
	if err != nil {
		return err
	}
	cfg.SetBootstrapPeers(bootPeers)

	// announce on 192.168.0.0
	cfg.Addresses.Announce = []string{"/ip4/192.168.0.0/ipcidr/16"}
	cfg.Ipns.ResolveCacheSize = 2048
	marshaledCfg, err := cg.Marshal(cfg)
	if err != nil {
		return err
	}

	outputFile, err := os.Create(ConfigFilePath)
	_, err = outputFile.Write(marshaledCfg)
	if err != nil {
		return err
	}
	outputSwarmKeyFile, err := os.Create(SwarmKeyPath)
	if err != nil {
		return err
	}
	swarmKey, err := genererateSwarmKey()
	if err != nil {
		return err
	}

	_, err = outputSwarmKeyFile.Write([]byte(swarmKey))
	if err != nil {
		return err
	}
	pcm.Config = cfg
	return nil
}

func genererateSwarmKey() (string, error) {
	var output string
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return "", err
	}
	keyEncoded := hex.EncodeToString(key)
	output = fmt.Sprintf("/key/swarm/psk/1.0.0/\n/base16/\n%s", keyEncoded)
	return output, nil
}
