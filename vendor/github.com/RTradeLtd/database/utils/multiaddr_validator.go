package utils

import (
	au "github.com/ipfs/go-ipfs-addr"
	ma "github.com/multiformats/go-multiaddr"
)

/*
Utilities that allow the user to validate multiaddr formatted addresses
*/

// GenerateMultiAddrFromString is used to take a string, and convert it to a multiformat based address
func GenerateMultiAddrFromString(addr string) (ma.Multiaddr, error) {
	var maddr ma.Multiaddr
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return nil, err
	}
	return maddr, nil
}

// ParseMultiAddrForIPFSPeer is used to parse a multiaddress to determine whether its a valid ipfs address
func ParseMultiAddrForIPFSPeer(address ma.Multiaddr) (bool, error) {
	protocols := address.Protocols()
	for _, v := range protocols {
		if v.Name == "ipfs" || v.Name == "p2p" {
			return true, nil
		}
	}
	return false, nil
}

// ParsePeerIDFromIPFSMultiAddr is used to parse a multiaddress and extract the IPFS peer id
func ParsePeerIDFromIPFSMultiAddr(address ma.Multiaddr) (string, error) {
	parsed, err := au.ParseMultiaddr(address)
	if err != nil {
		return "", err
	}
	pretty := parsed.ID().Pretty()
	return pretty, nil
}
