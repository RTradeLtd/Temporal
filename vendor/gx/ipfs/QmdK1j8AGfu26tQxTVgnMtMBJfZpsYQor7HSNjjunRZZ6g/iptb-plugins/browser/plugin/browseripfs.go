package main

import (
	testbedi "gx/ipfs/QmckeQ2zrYLAXoSHYTGn5BDdb22BqbUoHEHm8KZ9YWRxd1/iptb/testbed/interfaces"
	plugin "gx/ipfs/QmdK1j8AGfu26tQxTVgnMtMBJfZpsYQor7HSNjjunRZZ6g/iptb-plugins/browser"
)

var PluginName string
var NewNode testbedi.NewNodeFunc

func init() {
	PluginName = plugin.PluginName
	NewNode = plugin.NewNode
}
