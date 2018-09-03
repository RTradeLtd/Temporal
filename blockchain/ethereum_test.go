package blockchain_test

import (
	"testing"

	"github.com/RTradeLtd/Temporal/blockchain"
	"github.com/RTradeLtd/Temporal/config"
)

var configPath = "/home/solidity/config.json"

func TestGenerateEthereumConnectionManager(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	_, err = blockchain.GenerateEthereumConnectionManager(cfg, "infura")
	if err != nil {
		t.Fatal(err)
	}
}
