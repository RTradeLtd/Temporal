package commands

import (
	"context"
	"fmt"
	"path"
	"time"

	"github.com/pkg/errors"
	cli "github.com/urfave/cli"

	"github.com/ipfs/iptb/testbed"
)

var ConnectCmd = cli.Command{
	Category:  "CORE",
	Name:      "connect",
	Usage:     "connect sets of nodes together (or all)",
	ArgsUsage: "[nodes] [nodes]",
	Description: `
The connect command allows for connecting sets of nodes together.

Every node listed in the first set, will try to connect to every node
listed in the second set.

There are three variants of the command. It can accept no arugments,
a single argument, or two arguments. The no argument and single argument
expands out to the two argument usage.

$ iptb connect             => iptb connect [0-C] [0-C]
$ iptb connect [n-m]       => iptb connect [n-m] [n-m]
$ iptb connect [n-m] [i-k]

Sets of nodes can be expressed in the following ways

INPUT         EXPANDED
0             0
[0]           0
[0-4]         0,1,2,3,4
[0,2-4]       0,2,3,4
[2-4,0]       2,3,4,0
[0,2,4]       0,2,4
`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "timeout",
			Usage: "timeout on the command",
			Value: "30s",
		},
	},
	Action: func(c *cli.Context) error {
		flagRoot := c.GlobalString("IPTB_ROOT")
		flagTestbed := c.GlobalString("testbed")
		flagQuiet := c.GlobalBool("quiet")
		flagTimeout := c.String("timeout")

		timeout, err := time.ParseDuration(flagTimeout)
		if err != nil {
			return err
		}

		tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))
		args := c.Args()

		var results []Result
		switch c.NArg() {
		case 0:
			nodes, err := tb.Nodes()
			if err != nil {
				return err
			}

			fromto, err := parseRange(fmt.Sprintf("[0-%d]", len(nodes)-1))
			if err != nil {
				return err
			}

			results, err = connectNodes(tb, fromto, fromto, timeout)
			if err != nil {
				return err
			}
		case 1:
			fromto, err := parseRange(args[0])
			if err != nil {
				return err
			}

			results, err = connectNodes(tb, fromto, fromto, timeout)
			if err != nil {
				return err
			}
		case 2:
			from, err := parseRange(args[0])
			if err != nil {
				return err
			}

			to, err := parseRange(args[1])
			if err != nil {
				return err
			}

			results, err = connectNodes(tb, from, to, timeout)
			if err != nil {
				return err
			}
		default:
			return NewUsageError("connet accepts between 0 and 2 arguments")
		}

		return buildReport(results, flagQuiet)
	},
}

func connectNodes(tb testbed.BasicTestbed, from, to []int, timeout time.Duration) ([]Result, error) {
	var results []Result

	nodes, err := tb.Nodes()
	if err != nil {
		return results, err
	}

	for _, f := range from {
		for _, t := range to {
			if f == t {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			err = nodes[f].Connect(ctx, nodes[t])

			results = append(results, Result{
				Node:   f,
				Output: nil,
				Error:  errors.Wrapf(err, "node[%d] => node[%d]", f, t),
			})
		}
	}

	return results, nil
}
