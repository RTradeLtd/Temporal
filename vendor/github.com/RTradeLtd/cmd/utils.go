package cmd

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/RTradeLtd/config"
)

func run(commands map[string]Cmd, cfg config.TemporalConfig,
	flags map[string]string, args []string, baseOptions *flag.FlagSet) (noop bool) {

	// find command
	c, ok := commands[args[0]]
	if !ok {
		return true
	}

	// parse options
	if c.Options != nil && len(args) > 0 {
		if err := c.Options.Parse(args[1:]); err != nil {
			return true
		}
		args = c.Options.Args()
		fmt.Printf("%v\n", args)
	}

	// check for action and children
	if c.Action == nil && (c.Children == nil || len(c.Children) == 0) {
		return true
	} else if c.Action != nil {
		// parse arguments
		if c.Args != nil {
			if len(args)-1 != len(c.Args) {
				return true
			}
			for pos, name := range c.Args {
				flags[name] = args[pos+1]
			}
		}
		// execute action
		c.Action(cfg, flags)
	}

	// check for children and walk through them based on conditions
	if c.Children != nil && len(c.Children) > 0 {
		if len(args) > 1 {
			return run(c.Children, cfg, flags, args[1:], baseOptions)
		}
		if c.ChildRequired {
			if c.Description != "" {
				help(c.Description, strings.Join(os.Args, " "), nil, c.Children, c.Args, baseOptions)
			} else {
				help(c.Blurb, strings.Join(os.Args, " "), nil, c.Children, c.Args, baseOptions)
			}
		}
	}

	return false
}

func help(doc, exec string, args []string, cmds map[string]Cmd, requiredArgs []string, baseOptions *flag.FlagSet) {
	if args != nil && len(args) > 0 {
		c, found := cmds[args[0]]
		if found {
			exec += " " + args[0]
			if c.Description != "" {
				help(c.Description, exec, args[1:], c.Children, c.Args, baseOptions)
			} else {
				help(c.Blurb, exec, args[1:], c.Children, c.Args, baseOptions)
			}
		} else {
			fmt.Printf("command '%s %s' not found\n", exec, strings.Join(args, " "))
		}
		return
	}

	if doc == "" {
		doc = "no documentation available"
	}

	// print main doc
	println(doc)

	// flags must be placed after first exec
	var execParts = strings.SplitN(exec, " ", 2)
	exec = strings.Join(execParts, " [OPTIONS] ")

	// set required args
	var required string
	if requiredArgs != nil && len(requiredArgs) > 0 {
		required = "[" + strings.Join(requiredArgs, "] [") + "]"
	}

	// print child command documentation if there are any
	if len(cmds) > 0 {
		fmt.Printf(`
USAGE:

  %s [COMMAND]

COMMANDS:

`, exec)
		// calculate longest name
		var sortedCommands = make([]string, 0)
		var longestCmdNameLen = 0
		for name := range cmds {
			if cmds[name].Hidden {
				continue
			}

			sortedCommands = append(sortedCommands, name)
			if len(name) > longestCmdNameLen {
				longestCmdNameLen = len(name)
			}
		}
		sort.Strings(sortedCommands)

		// print help text for each command
		for _, name := range sortedCommands {
			dividerSpace := ""
			for i := 0; i < longestCmdNameLen-len(name); i++ {
				dividerSpace += " "
			}
			fmt.Printf("  %s%s  %s\n", name, dividerSpace, cmds[name].Blurb)
		}
	} else {
		fmt.Printf(`
USAGE:
		
  %s %s
`, exec, required)
	}

	// print help text for flags
	println(`
OPTIONS:
`)
	var noOpts = true
	if baseOptions != nil {
		baseOptions.PrintDefaults()
		noOpts = false
	}
	var calls = strings.Split(exec, " ")
	for _, c := range calls {
		if cmds[c].Options != nil {
			cmds[c].Options.PrintDefaults()
			noOpts = false
		}
	}
	if noOpts {
		println("  none")
	}
}
