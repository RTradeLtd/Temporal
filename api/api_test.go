package api

import (
	"os"
	"testing"

	"github.com/RTradeLtd/config"
	"github.com/gin-gonic/gin"
)

func Test_new(t *testing.T) {
	cfg, err := config.LoadConfig("../test/config.json")
	if err != nil {
		t.Fatal(err)
	}

	api, err := new(cfg, gin.New(), true, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	api.setupRoutes()
}
