package commands

import (
	"fmt"
	"io"
	"path"
	"strconv"

	cli "github.com/urfave/cli"

	"github.com/ipfs/iptb/testbed"
	"github.com/ipfs/iptb/testbed/interfaces"
)

var EventsCmd = cli.Command{
	Category:  "METRICS",
	Name:      "events",
	Usage:     "stream events from specified nodes (or all)",
	ArgsUsage: "[node]",
	Action: func(c *cli.Context) error {
		flagRoot := c.GlobalString("IPTB_ROOT")
		flagTestbed := c.GlobalString("testbed")

		if !c.Args().Present() {
			return NewUsageError("events takes exactly 1 argument")
		}

		i, err := strconv.Atoi(c.Args().First())
		if err != nil {
			return fmt.Errorf("parse err: %s", err)
		}

		tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))

		node, err := tb.Node(i)
		if err != nil {
			return err
		}

		mn, ok := node.(testbedi.Metric)
		if !ok {
			return fmt.Errorf("node does not implement metrics")
		}

		el, err := mn.Events()
		if err != nil {
			return err
		}

		_, err = io.Copy(c.App.Writer, el)
		return err
	},
}
