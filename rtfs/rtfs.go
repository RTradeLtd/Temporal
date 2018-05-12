package rtfs

import (
	ipfsapi "github.com/ipfs/go-ipfs-api"
)

type IpfsManager struct {
	Shell *ipfsapi.Shell
}

func Initialize() *IpfsManager {
	url := ""
	manager := IpfsManager{}
	manager.Shell = establishShellWithNode(url)
	return &manager
}

func establishShellWithNode(url string) *ipfsapi.Shell {
	if len(url) < 10 {
		shell := ipfsapi.NewLocalShell()
		return shell
	}
	shell := ipfsapi.NewShell(url)
	return shell
}
