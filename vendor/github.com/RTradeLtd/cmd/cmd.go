package cmd

import (
	"github.com/RTradeLtd/config"
)

// Cmd declares a command for a Temporal application.
type Cmd struct {
	Blurb         string
	Description   string
	Hidden        bool
	PreRun        bool
	ChildRequired bool

	Action   func(cfg config.TemporalConfig, flags map[string]string)
	Children map[string]Cmd
}
