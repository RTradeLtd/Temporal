package utils

import (
	au "github.com/ipfs/go-ipfs-addr"
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

func ParseMultiAddrForIPFSPeer(address ma.Multiaddr) (bool, error) {
	protocols := address.Protocols()
	for _, v := range protocols {
		if v.Name == "ipfs" || v.Name == "p2p" {
			return true, nil
		}
	}
	return false, nil
}

func ParsePeerIDFromIPFSMultiAddr(address ma.Multiaddr) (string, error) {
	parsed, err := au.ParseMultiaddr(address)
	if err != nil {
		return "", err
	}
	pretty := parsed.ID().Pretty()
	return pretty, nil
}
