package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/RTradeLtd/config"
)

func run(commands map[string]Cmd, cfg config.TemporalConfig,
	flags map[string]string, args []string) (noop bool) {

	// find command
	c, ok := commands[args[0]]
	if !ok {
		return true
	}

	// check for action and children
	if c.Action == nil && (c.Children == nil || len(c.Children) == 0) {
		return true
	} else if c.Action != nil {
		c.Action(cfg, flags)
	}

	// check for children and walk through them based on conditions
	if c.Children != nil && len(c.Children) > 0 {
		if len(args) > 1 {
			return run(c.Children, cfg, flags, args[1:])
		}
		if c.ChildRequired {
			if c.Description != "" {
				help(c.Description, strings.Join(os.Args, " "), nil, c.Children)
			} else {
				help(c.Blurb, strings.Join(os.Args, " "), nil, c.Children)
			}
		}
	}

	return false
}

func help(doc, exec string, args []string, cmds map[string]Cmd) {
	if args != nil && len(args) > 0 {
		c, found := cmds[args[0]]
		if found {
			exec += " " + args[0]
			if c.Description != "" {
				help(c.Description, exec, args[1:], c.Children)
			} else {
				help(c.Blurb, exec, args[1:], c.Children)
			}
		} else {
			fmt.Printf("command %s %s not found\n", exec, strings.Join(args, " "))
		}
		return
	}

	if doc == "" {
		doc = "no documentation available"
	}

	if cmds == nil || len(cmds) == 0 {
		println(doc)
		return
	}

	fmt.Printf(`%s

usage:

	%s [command]

commands:

`, doc, exec)

	// calculate longest name
	sortedCommands := make([]string, 0)
	longestCmdNameLen := 0
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
		fmt.Printf("	%s%s  %s\n", name, dividerSpace, cmds[name].Blurb)
	}
}
