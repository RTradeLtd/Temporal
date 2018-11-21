package commands

import (
	"fmt"
	"path"
	"strconv"

	cli "github.com/urfave/cli"

	"github.com/ipfs/iptb/testbed"
	"github.com/ipfs/iptb/testbed/interfaces"
)

var MetricCmd = cli.Command{
	Category:  "METRICS",
	Name:      "metric",
	Usage:     "get metric from node",
	ArgsUsage: "<node> [metric]",
	Action: func(c *cli.Context) error {
		if c.NArg() == 1 {
			return metricList(c)
		}

		if c.NArg() == 2 {
			return metricGet(c)
		}

		return NewUsageError("metric takes 1 or 2 arguments only")
	},
}

func metricList(c *cli.Context) error {
	flagRoot := c.GlobalString("IPTB_ROOT")
	flagTestbed := c.GlobalString("testbed")

	i, err := strconv.Atoi(c.Args().First())
	if err != nil {
		return fmt.Errorf("parse err: %s", err)
	}

	tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))

	node, err := tb.Node(i)
	if err != nil {
		return err
	}

	metricNode, ok := node.(testbedi.Metric)
	if !ok {
		return fmt.Errorf("node does not implement metrics")
	}

	metricList := metricNode.GetMetricList()
	for _, m := range metricList {
		desc, err := metricNode.GetMetricDesc(m)
		if err != nil {
			return fmt.Errorf("error getting metric description: %s", err)
		}

		fmt.Fprintf(c.App.Writer, "\t%s: %s\n", m, desc)
	}

	return nil
}

func metricGet(c *cli.Context) error {
	flagRoot := c.GlobalString("IPTB_ROOT")
	flagTestbed := c.GlobalString("testbed")

	argNode := c.Args()[0]
	argMetric := c.Args()[1]

	i, err := strconv.Atoi(argNode)
	if err != nil {
		return fmt.Errorf("parse err: %s", err)
	}

	tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))

	node, err := tb.Node(i)
	if err != nil {
		return err
	}

	metricNode, ok := node.(testbedi.Metric)
	if !ok {
		return fmt.Errorf("node does not implement metrics")
	}

	value, err := metricNode.Metric(argMetric)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(c.App.Writer, "%s\n", value)

	return err
}
