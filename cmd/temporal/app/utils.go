package app

import (
	"fmt"
	"sort"
	"strings"
)

func help(doc, exec string, args []string, cmds map[string]Cmd) {
	if len(args) > 0 {
		c, found := cmds[args[0]]
		if found {
			exec += " " + args[0]
			help(c.Description, exec, args[1:], c.Children)
		} else {
			println("command %s %s not found", exec, strings.Join(args, " "))
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
