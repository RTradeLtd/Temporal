package utils

import (
	au "github.com/ipfs/go-ipfs-addr"
	ma "github.com/multiformats/go-multiaddr"
)

func GenerateMultiAddrFromString(addr string) (ma.Multiaddr, error) {
	var maddr ma.Multiaddr
	maddr, err := ma.NewMultiaddr(addr)
	if err != nil {
		return nil, err
	}
	return maddr, nil
}

func ParsePeerIDFromIPFSMultiAddr(address ma.Multiaddr) (string, error) {
	parsed, err := au.ParseMultiaddr(address)
	if err != nil {
		return "", err
	}
	pretty := parsed.ID().Pretty()
	return pretty, nil
}
