package bstest

import (
	. "gx/ipfs/QmdHqV7L4bpmMtEXVCrgn8RN6CXqMr3aUeogSkXbJGRtwk/go-blockservice"

	delay "gx/ipfs/QmRJVNatYJwTAHgdSM1Xef9QVQ1Ch3XHdmcrykjP5Y4soL/go-ipfs-delay"
	mockrouting "gx/ipfs/QmZdn8S4FLTfDrmLZb7JoLkrRvTYnyuMWEG6ZGZ3YKwEiK/go-ipfs-routing/mock"
	bitswap "gx/ipfs/Qmd8rU7X3VZzsgPnf2LSGUFu35zizYKajzXTRuHMUMqYJQ/go-bitswap"
	tn "gx/ipfs/Qmd8rU7X3VZzsgPnf2LSGUFu35zizYKajzXTRuHMUMqYJQ/go-bitswap/testnet"
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
