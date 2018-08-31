package bstest

import (
	. "gx/ipfs/QmbSB9Uh3wVgmiCb1fAb8zuC3qAE6un4kd1jvatUurfAmB/go-blockservice"

	delay "gx/ipfs/QmRJVNatYJwTAHgdSM1Xef9QVQ1Ch3XHdmcrykjP5Y4soL/go-ipfs-delay"
	bitswap "gx/ipfs/QmTtmrK4iiM3MxWNA3pvbM9ekQiGZAiFyo57GP8B9FFgtz/go-bitswap"
	tn "gx/ipfs/QmTtmrK4iiM3MxWNA3pvbM9ekQiGZAiFyo57GP8B9FFgtz/go-bitswap/testnet"
	mockrouting "gx/ipfs/Qmd45r5jHr1PKMNQqifnbZy1ZQwHdtXUDJFamUEvUJE544/go-ipfs-routing/mock"
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
