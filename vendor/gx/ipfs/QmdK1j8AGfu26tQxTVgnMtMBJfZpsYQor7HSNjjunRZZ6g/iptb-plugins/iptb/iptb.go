package main

import (
	"fmt"
	"os"

	cli "gx/ipfs/QmckeQ2zrYLAXoSHYTGn5BDdb22BqbUoHEHm8KZ9YWRxd1/iptb/cli"
	testbed "gx/ipfs/QmckeQ2zrYLAXoSHYTGn5BDdb22BqbUoHEHm8KZ9YWRxd1/iptb/testbed"

	browser "gx/ipfs/QmdK1j8AGfu26tQxTVgnMtMBJfZpsYQor7HSNjjunRZZ6g/iptb-plugins/browser"
	docker "gx/ipfs/QmdK1j8AGfu26tQxTVgnMtMBJfZpsYQor7HSNjjunRZZ6g/iptb-plugins/docker"
	local "gx/ipfs/QmdK1j8AGfu26tQxTVgnMtMBJfZpsYQor7HSNjjunRZZ6g/iptb-plugins/local"
	localp2pd "gx/ipfs/QmdK1j8AGfu26tQxTVgnMtMBJfZpsYQor7HSNjjunRZZ6g/iptb-plugins/localp2pd"
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
		NewNode:     localp2pd.NewNode,
		GetAttrList: localp2pd.GetAttrList,
		GetAttrDesc: localp2pd.GetAttrDesc,
		PluginName:  localp2pd.PluginName,
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
