package api

import (
	"testing"

	"github.com/RTradeLtd/config"
)

func Test_Initialize(t *testing.T) {
	cfg, err := config.LoadConfig("../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Initialize(cfg, true); err != nil {
		t.Fatal(err)
	}
}
