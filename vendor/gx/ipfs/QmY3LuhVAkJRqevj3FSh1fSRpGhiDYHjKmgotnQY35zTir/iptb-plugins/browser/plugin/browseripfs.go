package main

import (
	plugin "gx/ipfs/QmY3LuhVAkJRqevj3FSh1fSRpGhiDYHjKmgotnQY35zTir/iptb-plugins/browser"
	testbedi "gx/ipfs/QmckeQ2zrYLAXoSHYTGn5BDdb22BqbUoHEHm8KZ9YWRxd1/iptb/testbed/interfaces"
)

var PluginName string
var NewNode testbedi.NewNodeFunc

func init() {
	PluginName = plugin.PluginName
	NewNode = plugin.NewNode
}
