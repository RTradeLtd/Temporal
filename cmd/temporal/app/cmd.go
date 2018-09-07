package app

import (
	"github.com/RTradeLtd/Temporal/config"
)

type Cmd struct {
	Blurb       string
	Description string
	Hidden      bool
	PreRun      bool

	Action   func(cfg config.TemporalConfig, flags map[string]string)
	Children map[string]Cmd
}
