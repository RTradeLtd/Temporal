# 🐢 cmd [![GoDoc](https://godoc.org/github.com/RTradeLtd/cmd?status.svg)](https://godoc.org/github.com/RTradeLtd/cmd) [![Build Status](https://travis-ci.com/RTradeLtd/cmd.svg?branch=master)](https://travis-ci.com/RTradeLtd/cmd) [![codecov](https://codecov.io/gh/RTradeLtd/cmd/branch/master/graph/badge.svg)](https://codecov.io/gh/RTradeLtd/cmd) [![Go Report Card](https://goreportcard.com/badge/github.com/RTradeLtd/cmd)](https://goreportcard.com/report/github.com/RTradeLtd/cmd)

Package cmd provides a microframework for building CLI tools integrated with [Temporal](https://github.com/RTradeLtd/Temporal) configuration.

It is extremely lightweight, with only a single dependency - package [`config`](https://github.com/RTradeLtd/config), which contains Temporal's configuration definitions.

## Usage

```go

import (
  "github.com/RTradeLtd/Temporal/cmd"
  "github.com/RTradeLtd/Temporal/config"
)

// define commands
var commands = map[string]cmd.Cmd{
  "api": cmd.Cmd{
    Blurb:       "start Temporal api server",
    Description: "Start the API service used to interact with Temporal. Run with DEBUG=true to enable debug messages.",
    Action: func(cfg config.TemporalConfig, args map[string]string) { /* ... */ },
  },
  "queue": cmd.Cmd{
    Blurb:         "execute commands for various queues",
    Description:   "Interact with Temporal's various queue APIs",
    ChildRequired: true,
    Children: map[string]cmd.Cmd{
      "ipns-entry": cmd.Cmd{
        Blurb:       "IPNS entry creation queue",
        Description: "Listens to requests to create IPNS records",
        Action: func(cfg config.TemporalConfig, args map[string]string) { /* ... */ },
      },
    },
  },
}

// entrypoint
func main() {
  // create app
  temporal := cmd.New(commands, cmd.Config{
    Name:     "Temporal",
    ExecName: "temporal",
    Version:  Version,
    Desc:     "Temporal is an easy-to-use interface into distributed and decentralized storage technologies for personal and enterprise use cases.",
  })

  // run no-config commands. exit if a command was executed
  if exit := temporal.PreRun(os.Args[1:]); exit == cmd.CodeOK {
    os.Exit(0)
  }

  // load config
  tCfg, _ := config.LoadConfig("path/to/config")

  // load arguments
  flags := map[string]string{ /* ... */ }

  // execute
  os.Exit(temporal.Run(*tCfg, flags, os.Args[1:]))
}
```
