package rtfs

import (
	ipfsapi "github.com/ipfs/go-ipfs-api"
)

type IpfsManager struct {
	Shell *ipfsapi.Shell
}

func Initialize() *IpfsManager {
	manager := IpfsManager{}
	manager.Shell = establishBackendLocalShell()
	return &manager
}

func establishBackendLocalShell() *ipfsapi.Shell {
	shell := ipfsapi.NewShell("localhost:25769")
	return shell
}

func establishShellWithNode(url string) *ipfsapi.Shell {
	if len(url) < 10 {
		shell := ipfsapi.NewLocalShell()
		return shell
	}
	shell := ipfsapi.NewShell(url)
	return shell
}
