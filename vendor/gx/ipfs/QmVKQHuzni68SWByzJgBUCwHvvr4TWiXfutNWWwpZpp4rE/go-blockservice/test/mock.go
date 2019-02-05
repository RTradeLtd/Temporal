package bstest

import (
	. "gx/ipfs/QmVKQHuzni68SWByzJgBUCwHvvr4TWiXfutNWWwpZpp4rE/go-blockservice"

	delay "gx/ipfs/QmUe1WCHkQaz4UeNKiHDUBV2T6i9prc3DniqyHPXyfGaUq/go-ipfs-delay"
	mockrouting "gx/ipfs/QmVZ6cQXHoTQja4oo9GhhHZi7dThi4x98mRKgGtKnTy37u/go-ipfs-routing/mock"
	bitswap "gx/ipfs/QmYokQouMdEuZjNQop75Bwm6ZV9CxJDcxvZHeSy4Ttzrtp/go-bitswap"
	tn "gx/ipfs/QmYokQouMdEuZjNQop75Bwm6ZV9CxJDcxvZHeSy4Ttzrtp/go-bitswap/testnet"
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
