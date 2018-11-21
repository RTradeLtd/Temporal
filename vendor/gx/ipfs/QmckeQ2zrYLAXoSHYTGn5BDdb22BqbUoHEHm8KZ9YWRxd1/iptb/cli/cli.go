package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	cli "github.com/urfave/cli"

	"github.com/ipfs/iptb/commands"
	"github.com/ipfs/iptb/testbed"
)

func loadPlugins(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	plugs, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range plugs {
		plg, err := testbed.LoadPlugin(path.Join(dir, f.Name()))

		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			continue
		}

		overloaded, err := testbed.RegisterPlugin(*plg, false)
		if overloaded {
			fmt.Fprintf(os.Stderr, "overriding built in plugin %s with %s\n", plg.PluginName, path.Join(dir, f.Name()))
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func NewCli() *cli.App {
	app := cli.NewApp()
	app.Usage = "iptb is a tool for managing test clusters of libp2p nodes"
	app.Version = "2.0.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "testbed",
			Value:  "default",
			EnvVar: "IPTB_TESTBED",
			Usage:  "Name of testbed to use under IPTB_ROOT",
		},
		cli.BoolFlag{
			Name:  "quiet",
			Usage: "Suppresses extra output from iptb",
		},
		cli.StringFlag{
			Name:   "IPTB_ROOT",
			EnvVar: "IPTB_ROOT",
			Hidden: true,
		},
	}
	app.Before = func(c *cli.Context) error {
		flagRoot := c.GlobalString("IPTB_ROOT")

		if len(flagRoot) == 0 {
			home := os.Getenv("HOME")
			if len(home) == 0 {
				return fmt.Errorf("environment variable HOME not set")
			}

			flagRoot = path.Join(home, "testbed")
		} else {
			var err error

			flagRoot, err = filepath.Abs(flagRoot)
			if err != nil {
				return err
			}
		}

		c.Set("IPTB_ROOT", flagRoot)

		return loadPlugins(path.Join(flagRoot, "plugins"))
	}
	app.Commands = []cli.Command{
		commands.AutoCmd,
		commands.TestbedCmd,

		commands.InitCmd,
		commands.StartCmd,
		commands.StopCmd,
		commands.RestartCmd,
		commands.RunCmd,
		commands.ConnectCmd,
		commands.ShellCmd,

		commands.AttrCmd,

		commands.LogsCmd,
		commands.EventsCmd,
		commands.MetricCmd,
	}

	// https://github.com/urfave/cli/issues/736
	// Currently unreleased
	/*
		app.ExitErrHandler = func(c *cli.Context, err error) {
			switch err.(type) {
			case *commands.UsageError:
				fmt.Fprintf(c.App.ErrWriter, "%s\n\n", err)
				cli.ShowCommandHelpAndExit(c, c.Command.Name, 1)
			default:
				cli.HandleExitCoder(err)
			}
		}
	*/

	app.ErrWriter = os.Stderr
	app.Writer = os.Stdout

	return app
}
