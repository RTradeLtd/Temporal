package commands

import (
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"

	cli "github.com/urfave/cli"

	"github.com/ipfs/iptb/testbed"
	"github.com/ipfs/iptb/testbed/interfaces"
)

var LogsCmd = cli.Command{
	Category:  "METRICS",
	Name:      "logs",
	Usage:     "show logs from specified nodes (or all)",
	ArgsUsage: "[nodes]",
	Flags: []cli.Flag{
		cli.BoolTFlag{
			Name:  "err, e",
			Usage: "show stderr stream",
		},
		cli.BoolTFlag{
			Name:  "out, o",
			Usage: "show stdout stream",
		},
	},
	Action: func(c *cli.Context) error {
		flagRoot := c.GlobalString("IPTB_ROOT")
		flagTestbed := c.GlobalString("testbed")
		flagQuiet := c.GlobalBool("quiet")
		flagErr := c.BoolT("err")
		flagOut := c.BoolT("out")

		tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))
		nodes, err := tb.Nodes()
		if err != nil {
			return err
		}

		nodeRange := c.Args().First()

		if nodeRange == "" {
			nodeRange = fmt.Sprintf("[0-%d]", len(nodes)-1)
		}

		list, err := parseRange(nodeRange)
		if err != nil {
			return err
		}

		runCmd := func(node testbedi.Core) (testbedi.Output, error) {
			metricNode, ok := node.(testbedi.Metric)
			if !ok {
				return nil, fmt.Errorf("node does not implement metrics")
			}

			stdout := ioutil.NopCloser(strings.NewReader(""))
			stderr := ioutil.NopCloser(strings.NewReader(""))

			if flagOut {
				var err error
				stdout, err = metricNode.StdoutReader()
				if err != nil {
					return nil, err
				}
			}

			if flagErr {
				var err error
				stderr, err = metricNode.StderrReader()
				if err != nil {
					return nil, err
				}
			}

			return NewOutput(stdout, stderr), nil
		}

		results, err := mapWithOutput(list, nodes, runCmd)
		if err != nil {
			return err
		}

		return buildReport(results, flagQuiet)
	},
}

func NewOutput(stdout, stderr io.ReadCloser) testbedi.Output {
	return &Output{
		stdout: stdout,
		stderr: stderr,
	}
}

type Output struct {
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func (o *Output) Args() []string {
	return []string{}
}

func (o *Output) Error() error {
	return nil
}
func (o *Output) ExitCode() int {
	return 0
}

func (o *Output) Stdout() io.ReadCloser {
	return o.stdout
}

func (o *Output) Stderr() io.ReadCloser {
	return o.stderr
}
