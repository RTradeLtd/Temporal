package main

import (
	"fmt"
	"os"

	cli "gx/ipfs/QmckeQ2zrYLAXoSHYTGn5BDdb22BqbUoHEHm8KZ9YWRxd1/iptb/cli"
	testbed "gx/ipfs/QmckeQ2zrYLAXoSHYTGn5BDdb22BqbUoHEHm8KZ9YWRxd1/iptb/testbed"

	browser "gx/ipfs/QmfAX8Pb621RhPM24wnmA1w9xewkXFBBAzJnzyi5AHtEyD/iptb-plugins/browser"
	docker "gx/ipfs/QmfAX8Pb621RhPM24wnmA1w9xewkXFBBAzJnzyi5AHtEyD/iptb-plugins/docker"
	local "gx/ipfs/QmfAX8Pb621RhPM24wnmA1w9xewkXFBBAzJnzyi5AHtEyD/iptb-plugins/local"
)

func init() {
	_, err := testbed.RegisterPlugin(testbed.IptbPlugin{
		From:        "<builtin>",
		NewNode:     local.NewNode,
		GetAttrList: local.GetAttrList,
		GetAttrDesc: local.GetAttrDesc,
		PluginName:  local.PluginName,
		BuiltIn:     true,
	}, false)

	if err != nil {
		panic(err)
	}

	_, err = testbed.RegisterPlugin(testbed.IptbPlugin{
		From:        "<builtin>",
		NewNode:     docker.NewNode,
		GetAttrList: docker.GetAttrList,
		GetAttrDesc: docker.GetAttrDesc,
		PluginName:  docker.PluginName,
		BuiltIn:     true,
	}, false)

	if err != nil {
		panic(err)
	}

	_, err = testbed.RegisterPlugin(testbed.IptbPlugin{
		From:       "<builtin>",
		NewNode:    browser.NewNode,
		PluginName: browser.PluginName,
		BuiltIn:    true,
	}, false)

	if err != nil {
		panic(err)
	}
}

func main() {
	cli := cli.NewCli()
	if err := cli.Run(os.Args); err != nil {
		fmt.Fprintf(cli.ErrWriter, "%s\n", err)
		os.Exit(1)
	}
}
