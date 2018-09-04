package main

import (
	"fmt"
	"sort"
	"strings"
)

func printNoOp(call []string, cmds map[string]cmd) {
	fmt.Printf("invalid invocation '%s'\n", strings.Join(call[:], " "))
	printHelp(cmds)
}

func printHelp(cmds map[string]cmd) {
	println(`usage:

	temporal [command]

commands:
`)

	// calculate longest name
	sortedCommands := make([]string, 0)
	longestCmdNameLen := 0
	for name := range cmds {
		if cmds[name].hidden {
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
		fmt.Printf("	%s%s  %s\n", name, dividerSpace, cmds[name].blurb)
	}

	// print built-ins
	println(`
	help      reveal help text for Temporal CLI
	version   report Temporal version`)
}
