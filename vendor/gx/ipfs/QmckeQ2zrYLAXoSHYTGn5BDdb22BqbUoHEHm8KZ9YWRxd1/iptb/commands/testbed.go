package commands

import (
	"context"
	"path"

	cli "github.com/urfave/cli"

	"github.com/ipfs/iptb/testbed"
)

var TestbedCmd = cli.Command{
	Name:  "testbed",
	Usage: "manage testbeds",
	Subcommands: []cli.Command{
		TestbedCreateCmd,
	},
}

var TestbedCreateCmd = cli.Command{
	Name:      "create",
	Usage:     "create testbed",
	ArgsUsage: "--type <type>",
	Flags: []cli.Flag{
		cli.IntFlag{
			Name:  "count",
			Usage: "number of nodes to initialize",
			Value: 1,
		},
		cli.BoolFlag{
			Name:  "force",
			Usage: "force overwrite of existing testbed",
		},
		cli.StringFlag{
			Name:  "type",
			Usage: "kind of nodes to initialize",
		},
		cli.StringSliceFlag{
			Name:  "attr",
			Usage: "specify addition attributes for nodes",
		},
		cli.BoolFlag{
			Name:  "init",
			Usage: "initialize after creation (like calling `init` after create)",
		},
	},
	Action: func(c *cli.Context) error {
		flagRoot := c.GlobalString("IPTB_ROOT")
		flagTestbed := c.GlobalString("testbed")
		flagType := c.String("type")
		flagInit := c.Bool("init")
		flagCount := c.Int("count")
		flagForce := c.Bool("force")
		flagAttrs := c.StringSlice("attr")

		attrs := parseAttrSlice(flagAttrs)
		tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))

		if err := testbed.AlreadyInitCheck(tb.Dir(), flagForce); err != nil {
			return err
		}

		specs, err := testbed.BuildSpecs(tb.Dir(), flagCount, flagType, attrs)
		if err != nil {
			return err
		}

		if err := testbed.WriteNodeSpecs(tb.Dir(), specs); err != nil {
			return err
		}

		if flagInit {
			nodes, err := tb.Nodes()
			if err != nil {
				return err
			}

			for _, n := range nodes {
				if _, err := n.Init(context.Background()); err != nil {
					return err
				}
			}
		}

		return nil
	},
}
