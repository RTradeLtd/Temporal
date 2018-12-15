package main

import (
	plugin "gx/ipfs/QmQwfyu1QgiKduoRqrztTA4yPWWx2dkntkuD5KzbXy77mc/iptb-plugins/browser"
	testbedi "gx/ipfs/QmckeQ2zrYLAXoSHYTGn5BDdb22BqbUoHEHm8KZ9YWRxd1/iptb/testbed/interfaces"
)

var PluginName string
var NewNode testbedi.NewNodeFunc

func init() {
	PluginName = plugin.PluginName
	NewNode = plugin.NewNode
}
