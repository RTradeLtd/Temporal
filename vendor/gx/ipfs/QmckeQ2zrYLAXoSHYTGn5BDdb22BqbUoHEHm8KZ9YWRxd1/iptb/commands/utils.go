package commands

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	cli "github.com/urfave/cli"

	"github.com/ipfs/iptb/testbed/interfaces"
)

// the flag terminator stops flag parsing, but it also swallowed if its the
// first argument into a command / subcommand. To find it, we have to look
// up to the parent command.
// iptb run 0 -- ipfs id => c.Args: 0 -- ipfs id
// iptb run -- ipfs id   => c.Args: ipfs id
func isTerminatorPresent(c *cli.Context) bool {
	argsParent := c.Parent().Args().Tail()
	argsSelf := c.Args()

	ls := len(argsSelf)
	lp := len(argsParent)

	term := lp - ls - 1

	if lp > ls && argsParent[term] == "--" {
		return true
	}

	return false
}

func parseAttrSlice(attrsraw []string) map[string]string {
	attrs := make(map[string]string)
	for _, attr := range attrsraw {
		parts := strings.Split(attr, ",")

		if len(parts) == 1 {
			attrs[parts[0]] = "true"
		} else {
			attrs[parts[0]] = strings.Join(parts[1:], ",")
		}
	}

	return attrs
}

func parseCommand(args []string, terminator bool) (string, []string) {
	if terminator {
		return "", args
	}

	if len(args) == 0 {
		return "", []string{}
	}

	if len(args) == 1 {
		return args[0], []string{}
	}

	if args[0] == "--" {
		return "", args[1:]
	}

	arguments := args[1:]

	if arguments[0] == "--" {
		arguments = arguments[1:]
	}

	return args[0], arguments
}

func parseRange(s string) ([]int, error) {
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		ranges := strings.Split(s[1:len(s)-1], ",")
		var out []int
		for _, r := range ranges {
			rng, err := expandDashRange(r)
			if err != nil {
				return nil, err
			}

			out = append(out, rng...)
		}
		return out, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}

	return []int{i}, nil
}

func expandDashRange(s string) ([]int, error) {
	parts := strings.Split(s, "-")
	if len(parts) == 1 {
		i, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		return []int{i}, nil
	}
	low, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}

	hi, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}

	var out []int
	for i := low; i <= hi; i++ {
		out = append(out, i)
	}
	return out, nil
}

type Result struct {
	Node   int
	Output testbedi.Output
	Error  error
}

type outputFunc func(testbedi.Core) (testbedi.Output, error)

func mapListWithOutput(ranges [][]int, nodes []testbedi.Core, fns []outputFunc) ([]Result, error) {
	var wg sync.WaitGroup
	var lk sync.Mutex
	var errs []error

	total := 0
	offsets := make([]int, len(ranges))
	for i, list := range ranges {
		offsets[i] = total
		total += len(list)
	}
	results := make([]Result, total)

	for i, list := range ranges {
		wg.Add(1)
		go func(i int, list []int) {
			defer wg.Done()
			results_i, err := mapWithOutput(list, nodes, fns[i])

			lk.Lock()
			defer lk.Unlock()

			if err != nil {
				errs = append(errs, err)
			}
			for j, result := range results_i {
				results[offsets[i]+j] = result
			}
		}(i, list)
		wg.Wait()
	}

	if len(errs) != 0 {
		return results, cli.NewMultiError(errs...)
	}

	return results, nil
}

func mapWithOutput(list []int, nodes []testbedi.Core, fn outputFunc) ([]Result, error) {
	var wg sync.WaitGroup
	var lk sync.Mutex
	results := make([]Result, len(list))

	if err := validRange(list, len(nodes)); err != nil {
		return results, err
	}

	for i, n := range list {
		wg.Add(1)
		go func(i, n int, node testbedi.Core) {
			defer wg.Done()
			out, err := fn(node)

			lk.Lock()
			defer lk.Unlock()

			results[i] = Result{
				Node:   n,
				Output: out,
				Error:  errors.Wrapf(err, "node[%d]", n),
			}
		}(i, n, nodes[n])
	}

	wg.Wait()

	return results, nil

}

func validRange(list []int, total int) error {
	max := 0
	for _, n := range list {
		if max < n {
			max = n
		}
	}

	if max >= total {
		return fmt.Errorf("Node range contains value (%d) outside of valid range [0-%d]", max, total-1)
	}

	return nil
}

func buildReport(results []Result, quiet bool) error {
	var errs []error

	for _, rs := range results {
		if rs.Error != nil {
			errs = append(errs, rs.Error)
		}

		if quiet {
			io.Copy(os.Stdout, rs.Output.Stdout())
			io.Copy(os.Stdout, rs.Output.Stderr())
			continue
		}

		if rs.Output != nil {
			fmt.Printf("node[%d] exit %d\n", rs.Node, rs.Output.ExitCode())
			if rs.Output.Error() != nil {
				fmt.Printf("%s", rs.Output.Error())
			}

			fmt.Println()

			io.Copy(os.Stdout, rs.Output.Stdout())
			io.Copy(os.Stdout, rs.Output.Stderr())

			fmt.Println()
		}

	}

	if len(errs) != 0 {
		return cli.NewMultiError(errs...)
	}

	return nil
}
