package main

import (
	plugin "gx/ipfs/QmTqLBwme9BusYWdACqL62NFb8WV2Q72gXLsQVfC7vmCr4/iptb-plugins/browser"
	testbedi "gx/ipfs/QmckeQ2zrYLAXoSHYTGn5BDdb22BqbUoHEHm8KZ9YWRxd1/iptb/testbed/interfaces"
)

var PluginName string
var NewNode testbedi.NewNodeFunc

func init() {
	PluginName = plugin.PluginName
	NewNode = plugin.NewNode
}
