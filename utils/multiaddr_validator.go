package utils

import (
	"fmt"

	ma "github.com/multiformats/go-multiaddr"
)

/*
Utilities that allow the user to validate multiaddr formatted addresses

*/

func GenerateMultiAddrFromString(addr string) (ma.Multiaddr, error) {
	var maddr ma.Multiaddr
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return nil, err
	}
	return maddr, nil
}

func ParseMultiAddrForBootstrap(address ma.Multiaddr) (bool, error) {
	protocols := address.Protocols()
	for _, v := range protocols {
		fmt.Println(v.Name)
		if v.Name == "ipfs" || v.Name == "p2p" {
			return true, nil
		}
	}
	return false, nil
}
