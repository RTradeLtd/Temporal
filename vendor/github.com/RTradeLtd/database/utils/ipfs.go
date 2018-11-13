package utils

import (
	au "github.com/ipsn/go-ipfs/gxlibs/github.com/ipfs/go-ipfs-addr"
	ma "github.com/ipsn/go-ipfs/gxlibs/github.com/multiformats/go-multiaddr"
)

func GenerateMultiAddrFromString(addr string) (ma.Multiaddr, error) {
	return ma.NewMultiaddr(addr)
}

func ParsePeerIDFromIPFSMultiAddr(address ma.Multiaddr) (string, error) {
	parsed, err := au.ParseMultiaddr(address)
	if err != nil {
		return "", err
	}
	return parsed.ID().Pretty(), nil
}
