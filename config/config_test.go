package config_test

import (
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/config"
)

var configPath = "../test/config.json"

func TestLoadConfig(t *testing.T) {
	if _, err := config.LoadConfig(configPath); err != nil {
		t.Fatal(err)
	}
}

func TestGenerateConfig(t *testing.T) {
	testconf := "./testconfig.json"
	defer os.Remove(testconf)
	if err := config.GenerateConfig(testconf); err != nil {
		t.Fatal(err)
	}
	if _, err := config.LoadConfig(testconf); err != nil {
		t.Fatal(err)
	}
}
