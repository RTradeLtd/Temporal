package bstest

import (
	. "gx/ipfs/QmPoh3SrQzFBWtdGK6qmHDV4EanKR6kYPj4DD3J2NLoEmZ/go-blockservice"

	bitswap "gx/ipfs/QmUYXFM46WgGs5AScfL4FSZXa9p5nAhddueyM5auAVZGCQ/go-bitswap"
	tn "gx/ipfs/QmUYXFM46WgGs5AScfL4FSZXa9p5nAhddueyM5auAVZGCQ/go-bitswap/testnet"
	delay "gx/ipfs/QmUe1WCHkQaz4UeNKiHDUBV2T6i9prc3DniqyHPXyfGaUq/go-ipfs-delay"
	mockrouting "gx/ipfs/QmdmWkx54g7VfVyxeG8ic84uf4G6Eq1GohuyKA3XDuJ8oC/go-ipfs-routing/mock"
)

// Mocks returns |n| connected mock Blockservices
func Mocks(n int) []BlockService {
	net := tn.VirtualNetwork(mockrouting.NewServer(), delay.Fixed(0))
	sg := bitswap.NewTestSessionGenerator(net)

	instances := sg.Instances(n)

	var servs []BlockService
	for _, i := range instances {
		servs = append(servs, New(i.Blockstore(), i.Exchange))
	}
	return servs
}
