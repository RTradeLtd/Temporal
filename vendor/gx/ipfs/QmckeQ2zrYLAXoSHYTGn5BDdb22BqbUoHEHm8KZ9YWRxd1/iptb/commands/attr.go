package commands

import (
	"fmt"
	"path"
	"strconv"

	cli "github.com/urfave/cli"

	"github.com/ipfs/iptb/testbed"
	"github.com/ipfs/iptb/testbed/interfaces"
)

var AttrCmd = cli.Command{
	Category: "ATTRIBUTES",
	Name:     "attr",
	Usage:    "get, set, list attributes",
	Subcommands: []cli.Command{
		AttrSetCmd,
		AttrGetCmd,
		AttrListCmd,
	},
}

var AttrSetCmd = cli.Command{
	Name:      "set",
	Usage:     "set an attribute for a node",
	ArgsUsage: "<node> <attr> <value>",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "save",
			Usage: "saves attribute value to nodespec",
		},
	},
	Action: func(c *cli.Context) error {
		flagRoot := c.GlobalString("IPTB_ROOT")
		flagTestbed := c.GlobalString("testbed")
		flagSave := c.Bool("save")

		if c.NArg() != 3 {
			return NewUsageError("set takes exactly 3 argument")
		}

		argNode := c.Args()[0]
		argAttr := c.Args()[1]
		argValue := c.Args()[2]

		i, err := strconv.Atoi(argNode)
		if err != nil {
			return fmt.Errorf("parse err: %s", err)
		}

		tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))

		node, err := tb.Node(i)
		if err != nil {
			return err
		}

		attrNode, ok := node.(testbedi.Attribute)
		if !ok {
			return fmt.Errorf("node does not implement attributes")
		}

		if err := attrNode.SetAttr(argAttr, argValue); err != nil {
			return err
		}

		if flagSave {
			specs, err := tb.Specs()
			if err != nil {
				return err
			}

			specs[i].SetAttr(argAttr, argValue)

			if err := testbed.WriteNodeSpecs(tb.Dir(), specs); err != nil {
				return err
			}
		}

		return nil
	},
}

var AttrGetCmd = cli.Command{
	Name:      "get",
	Usage:     "get an attribute for a node",
	ArgsUsage: "<node> <attr>",
	Action: func(c *cli.Context) error {
		flagRoot := c.GlobalString("IPTB_ROOT")
		flagTestbed := c.GlobalString("testbed")

		if c.NArg() != 2 {
			return NewUsageError("get takes exactly 2 argument")
		}

		argNode := c.Args()[0]
		argAttr := c.Args()[1]

		i, err := strconv.Atoi(argNode)
		if err != nil {
			return fmt.Errorf("parse err: %s", err)
		}

		tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))

		node, err := tb.Node(i)
		if err != nil {
			return err
		}

		attrNode, ok := node.(testbedi.Attribute)
		if !ok {
			return fmt.Errorf("node does not implement attributes")
		}

		value, err := attrNode.Attr(argAttr)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(c.App.Writer, "%s\n", value)

		return err
	},
}

var AttrListCmd = cli.Command{
	Name:      "list",
	Usage:     "list attributes available for a node",
	ArgsUsage: "<node>",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "type",
			Usage: "look up attributes for node type",
		},
	},
	Action: func(c *cli.Context) error {
		flagRoot := c.GlobalString("IPTB_ROOT")
		flagTestbed := c.GlobalString("testbed")
		flagType := c.String("type")

		if !c.Args().Present() && len(flagType) == 0 {
			return NewUsageError("specify a node, or a type")
		}

		if c.Args().Present() {
			i, err := strconv.Atoi(c.Args().First())
			if err != nil {
				return fmt.Errorf("parse err: %s", err)
			}

			tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))

			spec, err := tb.Spec(i)
			if err != nil {
				return err
			}

			flagType = spec.Type
		}

		plg, ok := testbed.GetPlugin(flagType)
		if !ok {
			return fmt.Errorf("Unknown plugin %s", flagType)
		}

		attrList := plg.GetAttrList()
		for _, a := range attrList {
			desc, err := plg.GetAttrDesc(a)
			if err != nil {
				return fmt.Errorf("error getting attribute description: %s", err)
			}

			fmt.Fprintf(c.App.Writer, "\t%s: %s\n", a, desc)
		}

		return nil
	},
}
