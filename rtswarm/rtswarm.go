package rtswarm

import (
	"github.com/ethereum/go-ethereum/swarm/api"
)

func GenDefaultConfig() *api.Config {
	return api.NewDefaultConfig()
}
