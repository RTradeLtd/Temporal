package config_test

import (
	"fmt"
	"testing"

	"github.com/RTradeLtd/Temporal/config"
)

// Change to your home dir
var configPath = "/home/solidity/config.json"

func TestConfig(t *testing.T) {
	cfg := config.LoadConfig(configPath)
	fmt.Printf("%+v\n", cfg.AWS)
}
