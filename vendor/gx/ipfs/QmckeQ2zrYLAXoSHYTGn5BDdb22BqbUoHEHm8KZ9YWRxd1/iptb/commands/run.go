package commands

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	cli "github.com/urfave/cli"

	"github.com/gxed/go-shellwords"
	"github.com/ipfs/iptb/testbed"
	"github.com/ipfs/iptb/testbed/interfaces"
)

var RunCmd = cli.Command{
	Category:  "CORE",
	Name:      "run",
	Usage:     "concurrently run command(s) on specified nodes (or all)",
	ArgsUsage: "[nodes] -- <command...>",
	Description: `
Commands may also be passed in via stdin or a pipe, e.g. the command

$ iptb run 0 -- echo "Running on node 0"

can be equivalently written as

$ iptb run <<CMD
    0 -- echo "Running on node 0"
  CMD

or

$ echo '0 -- echo "Running on node 0"' | iptb run

All lines starting with '#' will be ignored, which allows for comments:

$ iptb run <<CMD
    # print ipfs peers
    0 -- ipfs swarm peers
  CMD

Multiple commands may also be passed via stdin/pipe:

$ iptb run <<CMDS
    0     -- echo "Running on node 0"
    [0,1] -- echo "Running on nodes 0 and 1"
          -- echo "Running on all nodes"
  CMDS

Note that any single call to ` + "`iptb run`" + ` runs *all* commands concurrently. So,
in the above example, there is no guarantee as to the order in which the lines
are printed.
`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:   "terminator",
			Hidden: true,
		},
		cli.BoolFlag{
			Name:   "stdin",
			Hidden: true,
		},
	},
	Before: func(c *cli.Context) error {
		if c.NArg() == 0 {
			finfo, err := os.Stdin.Stat()
			if err != nil {
				return err
			}
			if finfo.Size() == 0 && finfo.Mode()&os.ModeNamedPipe == 0 {
				return fmt.Errorf("error: no command input and stdin is empty")
			}
			return c.Set("stdin", "true")
		}
		if present := isTerminatorPresent(c); present {
			return c.Set("terminator", "true")
		}
		return nil
	},
	Action: func(c *cli.Context) error {
		flagRoot := c.GlobalString("IPTB_ROOT")
		flagTestbed := c.GlobalString("testbed")
		flagQuiet := c.GlobalBool("quiet")

		tb := testbed.NewTestbed(path.Join(flagRoot, "testbeds", flagTestbed))
		nodes, err := tb.Nodes()
		if err != nil {
			return err
		}

		var reader io.Reader
		if c.IsSet("stdin") {
			reader = bufio.NewReader(os.Stdin)
		} else {
			var builder strings.Builder
			if c.IsSet("terminator") {
				builder.WriteString("-- ")
			}
			for i, arg := range c.Args() {
				builder.WriteString(strconv.Quote(arg))
				if i != c.NArg()-1 {
					builder.WriteString(" ")
				}
			}
			reader = strings.NewReader(builder.String())
		}

		var args [][]string
		scanner := bufio.NewScanner(reader)
		line := 1
		for scanner.Scan() {
			tokens, err := shellwords.Parse(scanner.Text())
			if err != nil {
				return fmt.Errorf("parse error on line %d: %s", line, err)
			}
			if strings.HasPrefix(tokens[0], "#") {
				continue
			}
			args = append(args, tokens)
			line++
		}

		ranges := make([][]int, len(args))
		runCmds := make([]outputFunc, len(args))
		for i, cmd := range args {
			nodeRange, tokens := parseCommand(cmd, false)
			if nodeRange == "" {
				nodeRange = fmt.Sprintf("[0-%d]", len(nodes)-1)
			}
			list, err := parseRange(nodeRange)
			if err != nil {
				return fmt.Errorf("could not parse node range %s", nodeRange)
			}
			ranges[i] = list

			runCmd := func(node testbedi.Core) (testbedi.Output, error) {
				return node.RunCmd(context.Background(), nil, tokens...)
			}
			runCmds[i] = runCmd
		}

		results, err := mapListWithOutput(ranges, nodes, runCmds)
		return buildReport(results, flagQuiet)
	},
}
