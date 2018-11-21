package commands

import (
	"context"
	"path"

	cli "github.com/urfave/cli"

	"github.com/ipfs/iptb/testbed"
	"github.com/ipfs/iptb/testbed/interfaces"
)

var AutoCmd = cli.Command{
	Name:  "auto",
	Usage: "create default testbed and initialize",
	Description: `
The auto command is a quick way to use iptb for simple configurations.

The auto command is similar to 'testbed create' except in a few ways

 - No attr options can be passed in
 - All nodes are initialize by default and ready to be started
 - An optional --start flag can be passed to start all nodes

The following two examples are equivalent

$ iptb testbed create -count 5 -type <type> -init
$ iptb auto           -count 5 -type <type>
`,
	ArgsUsage: "--type <type>",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "count",
			Usage: "number of nodes to initialize",
			Value: 1,
		},
		cli.BoolFlag{
			Name:  "force",
			Usage: "force overwrite of existing nodespecs",
		},
		cli.StringFlag{
			Name:  "type",
			Usage: "kind of nodes to initialize",
		},
		cli.BoolFlag{
			Name:  "start",
			Usage: "starts nodes immediately",
		},
	},
	Action: func(c *cli.Context) error {
		flagRoot := c.GlobalString("IPTB_ROOT")
		flagTestbed := c.GlobalString("testbed")
		flagQuiet := c.GlobalBool("quiet")
		flagType := c.String("type")
		flagStart := c.Bool("start")
		flagCount := c.Int("count")
		flagForce := c.Bool("force")

		tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))

		if err := testbed.AlreadyInitCheck(tb.Dir(), flagForce); err != nil {
			return err
		}

		specs, err := testbed.BuildSpecs(tb.Dir(), flagCount, flagType, nil)
		if err != nil {
			return err
		}

		if err := testbed.WriteNodeSpecs(tb.Dir(), specs); err != nil {
			return err
		}

		nodes, err := tb.Nodes()
		if err != nil {
			return err
		}

		var list []int
		for i, _ := range nodes {
			list = append(list, i)
		}

		runCmd := func(node testbedi.Core) (testbedi.Output, error) {
			return node.Init(context.Background())
		}

		results, err := mapWithOutput(list, nodes, runCmd)
		if err != nil {
			return err
		}

		if err := buildReport(results, flagQuiet); err != nil {
			return err
		}

		if flagStart {
			runCmd := func(node testbedi.Core) (testbedi.Output, error) {
				return node.Start(context.Background(), true)
			}

			results, err := mapWithOutput(list, nodes, runCmd)
			if err != nil {
				return err
			}

			if err := buildReport(results, flagQuiet); err != nil {
				return err
			}
		}

		return nil
	},
}
