package rtfs

import (
	ipfsapi "github.com/ipfs/go-ipfs-api"
)

type IpfsManager struct {
	Shell *ipfsapi.Shell
}

func Initialize() *IpfsManager {
	manager := IpfsManager{}
	manager.Shell = establishShellWithNode("")
	return &manager
}

func establishShellWithNode(url string) *ipfsapi.Shell {
	if url == "" {
		shell := ipfsapi.NewLocalShell()
		return shell
	}
	shell := ipfsapi.NewShell(url)
	return shell
}
