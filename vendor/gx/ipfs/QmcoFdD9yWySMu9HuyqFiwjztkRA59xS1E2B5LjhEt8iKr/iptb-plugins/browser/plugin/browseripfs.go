package main

import (
	testbedi "gx/ipfs/QmckeQ2zrYLAXoSHYTGn5BDdb22BqbUoHEHm8KZ9YWRxd1/iptb/testbed/interfaces"
	plugin "gx/ipfs/QmcoFdD9yWySMu9HuyqFiwjztkRA59xS1E2B5LjhEt8iKr/iptb-plugins/browser"
)

var PluginName string
var NewNode testbedi.NewNodeFunc

func init() {
	PluginName = plugin.PluginName
	NewNode = plugin.NewNode
}
