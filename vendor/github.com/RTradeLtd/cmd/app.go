package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/RTradeLtd/config"
)

const (
	// CodeOK indicates success
	CodeOK = 0

	// CodeNoOp indicates failure
	CodeNoOp = 1
)

// Config declares application settings
type Config struct {
	Name     string
	ExecName string
	Version  string
	Desc     string
}

// App is a Temporal command line app
type App struct {
	cfg  Config
	cmds map[string]Cmd
}

// New creates a new Temporal command line app
func New(cmds map[string]Cmd, cfg Config) *App {
	app := &App{cfg, cmds}
	app.cmds["help"] = Cmd{
		Blurb:       "show help text",
		Description: fmt.Sprintf("Display help text for %s or the given command", cfg.Name),
		Action:      func(config.TemporalConfig, map[string]string) { app.help() },
		PreRun:      true,
	}
	if cfg.Version != "" {
		cmds["version"] = Cmd{
			Blurb:       fmt.Sprintf("show %s version", cfg.Name),
			Description: fmt.Sprintf("Show the version of %s in use", cfg.Name),
			Action:      func(config.TemporalConfig, map[string]string) { app.version() },
			PreRun:      true,
		}
	}
	return app
}

// PreRun walks the command tree and executes commands marked as "PreRun" - this
// is useful for running commands that do not need configuration to be read,
// for example the built-in "help" command.
func (a *App) PreRun(args []string) int {
	if len(args) == 0 {
		a.help()
		return CodeOK
	}
	prerunCmds := make(map[string]Cmd)
	for exec, cmd := range a.cmds {
		if cmd.PreRun {
			prerunCmds[exec] = cmd
		}
	}

	if noop := run(prerunCmds, config.TemporalConfig{}, nil, args); noop {
		return CodeNoOp
	}
	return CodeOK
}

// Run walks the command tree and executes them as appropriate.
func (a *App) Run(cfg config.TemporalConfig, flags map[string]string, args []string) int {
	if len(args) == 0 {
		a.help()
		return CodeOK
	}

	if noop := run(a.cmds, cfg, flags, args); noop {
		a.noop(args)
		return CodeNoOp
	}
	return CodeOK
}

func (a *App) help() {
	if len(os.Args) > 2 {
		help(a.cfg.Desc, os.Args[0], os.Args[2:], a.cmds)
	} else {
		help(a.cfg.Desc, os.Args[0], []string{}, a.cmds)
	}
}

func (a *App) noop(args []string) {
	fmt.Printf("invalid invocation '%s %s'\n", a.cfg.ExecName, strings.Join(args[:], " "))
	fmt.Printf("\nUse '%s help [command]' to see CLI documentation.", a.cfg.ExecName)
}

func (a *App) version() {
	fmt.Printf("%s %s", a.cfg.Name, a.cfg.Version)
}
