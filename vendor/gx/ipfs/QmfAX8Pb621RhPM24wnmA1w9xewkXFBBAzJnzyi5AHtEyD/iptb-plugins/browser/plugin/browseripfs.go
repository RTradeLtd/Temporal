package main

import (
	testbedi "gx/ipfs/QmckeQ2zrYLAXoSHYTGn5BDdb22BqbUoHEHm8KZ9YWRxd1/iptb/testbed/interfaces"
	plugin "gx/ipfs/QmfAX8Pb621RhPM24wnmA1w9xewkXFBBAzJnzyi5AHtEyD/iptb-plugins/browser"
)

var PluginName string
var NewNode testbedi.NewNodeFunc

func init() {
	PluginName = plugin.PluginName
	NewNode = plugin.NewNode
}
