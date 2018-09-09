package config_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/config"
)

var configPath = "../test/config.json"

func TestConfig(t *testing.T) {
	if _, err := config.LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}
}
