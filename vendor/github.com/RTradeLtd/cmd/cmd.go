package cmd

import (
	"flag"

	"github.com/RTradeLtd/config"
)

// Cmd declares a command for a Temporal application.
type Cmd struct {
	Blurb         string
	Description   string
	Hidden        bool
	PreRun        bool
	ChildRequired bool
	Children      map[string]Cmd

	Args    []string      // names of required positional arguments. order is enforced
	Options *flag.FlagSet // command flags

	// flags include arguments loaded by Cmd.Args
	Action func(cfg config.TemporalConfig, flags map[string]string)
}
