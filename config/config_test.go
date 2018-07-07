package config_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/config"
)

// Change to your home dir
var configPath = "/home/solidity/config.json"

func TestConfig(t *testing.T) {
	_, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Printf("%+v\n", cfg.AWS)
	//fmt.Printf("%+v\n", cfg.MINIO)
}
