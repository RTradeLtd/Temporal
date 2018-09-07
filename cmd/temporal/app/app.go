package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/RTradeLtd/Temporal/config"
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
	}
	if cfg.Version != "" {
		cmds["version"] = Cmd{
			Blurb:       fmt.Sprintf("show %s version", cfg.Name),
			Description: fmt.Sprintf("Show the version of %s in use", cfg.Name),
			Action:      func(config.TemporalConfig, map[string]string) { app.version() },
		}
	}
	return app
}

// PreRun walks the command tree and executes commands marked as "PreRun" - this
// is useful for running commands that do not need configuration to be read,
// for example the built-in "help" command
func (a *App) PreRun(args []string) {
	if len(args) == 0 {
		a.help()
		os.Exit(0)
	}
	prerunCmds := make(map[string]Cmd)
	for exec, cmd := range a.cmds {
		if cmd.PreRun {
			prerunCmds[exec] = cmd
		}
	}

	noop := a.run(prerunCmds, config.TemporalConfig{}, nil, args)
	if noop {
		a.noop(args)
		os.Exit(1)
	}
	os.Exit(0)
}

// Run walks the command tree and executes them as appropriate.
func (a *App) Run(cfg config.TemporalConfig, flags map[string]string, args []string) {
	if len(args) == 0 {
		a.help()
		os.Exit(0)
	}

	noop := a.run(a.cmds, cfg, flags, args)
	if noop {
		a.noop(args)
		os.Exit(1)
	}
}

func (a *App) run(commands map[string]Cmd,
	cfg config.TemporalConfig, flags map[string]string, args []string) (noop bool) {
	c, ok := a.cmds[args[0]]
	if !ok {
		return true
	}
	if c.Action == nil {
		return true
	}
	c.Action(cfg, flags)
	if c.Children != nil {
		return a.run(c.Children, cfg, flags, args[1:])
	}
	return false
}

func (a *App) help() {
	help(a.cfg.Desc, os.Args[0], os.Args[2:], a.cmds)
}

func (a *App) noop(args []string) {
	fmt.Printf("invalid invocation '%s %s'\n", a.cfg.ExecName, strings.Join(args[:], " "))
	fmt.Printf("\nUse '%s help [command]' to see CLI documentation.", a.cfg.ExecName)
}

func (a *App) version() {
	fmt.Printf("%s %s", a.cfg.Name, a.cfg.Version)
}
