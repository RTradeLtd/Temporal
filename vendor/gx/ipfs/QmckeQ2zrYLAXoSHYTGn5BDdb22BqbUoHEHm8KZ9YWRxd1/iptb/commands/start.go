package commands

import (
	"context"
	"fmt"
	"path"

	cli "github.com/urfave/cli"

	"github.com/ipfs/iptb/testbed"
	"github.com/ipfs/iptb/testbed/interfaces"
)

var StartCmd = cli.Command{
	Category:  "CORE",
	Name:      "start",
	Usage:     "start specified nodes (or all)",
	ArgsUsage: "[nodes] -- [arguments...]",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "wait",
			Usage: "wait for nodes to start before returning",
		},
		cli.BoolFlag{
			Name:   "terminator",
			Hidden: true,
		},
	},
	Before: func(c *cli.Context) error {
		if present := isTerminatorPresent(c); present {
			return c.Set("terminator", "true")
		}

		return nil
	},
	Action: func(c *cli.Context) error {
		flagRoot := c.GlobalString("IPTB_ROOT")
		flagTestbed := c.GlobalString("testbed")
		flagQuiet := c.GlobalBool("quiet")
		flagWait := c.Bool("wait")

		tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))
		nodes, err := tb.Nodes()
		if err != nil {
			return err
		}

		nodeRange, args := parseCommand(c.Args(), c.IsSet("terminator"))

		if nodeRange == "" {
			nodeRange = fmt.Sprintf("[0-%d]", len(nodes)-1)
		}

		list, err := parseRange(nodeRange)
		if err != nil {
			return fmt.Errorf("could not parse node range %s", nodeRange)
		}

		runCmd := func(node testbedi.Core) (testbedi.Output, error) {
			return node.Start(context.Background(), flagWait, args...)
		}

		results, err := mapWithOutput(list, nodes, runCmd)
		if err != nil {
			return err
		}

		return buildReport(results, flagQuiet)
	},
}
