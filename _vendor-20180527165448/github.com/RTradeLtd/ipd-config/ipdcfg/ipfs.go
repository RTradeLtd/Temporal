package ipdcfg

import (
	ipfsapi "github.com/ipfs/go-ipfs-api"
)

func Initialize(nodeUrl string) *Config {
	config := Config{}
	config.Shell = connectToNetwork(nodeUrl)
	return &config
}

// ConnectToNetwork is used to establish a connection to the ifps network
func connectToNetwork(url string) *ipfsapi.Shell {
	switch url {
	case "":
		return ipfsapi.NewLocalShell()
	default:
		return ipfsapi.NewShell(url)
	}
}
