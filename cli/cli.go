package cli

import (
	ishell "gopkg.in/abiosoft/ishell.v2"
)

type CommandLine struct {
	Shell *ishell.Shell
}

// Initialize is used to init our command line app
func Initialize() {
	var cli CommandLine
	shell := ishell.New()
	cli.Shell = shell
	cli.SetupShell()
	cli.Shell.Run()
}

func (cl *CommandLine) SetupShell() {
	cl.Shell.Println("Temporal Command Line Interactive Shell")
	cl.Shell.Println("Version 0.0.5alpha")
}
