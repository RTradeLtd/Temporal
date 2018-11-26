package api

import (
	"os"
	"testing"
	"time"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/rtfs"
	"github.com/RTradeLtd/rtfs/krab"
	"github.com/gin-gonic/gin"
)

func Test_new(t *testing.T) {
	cfg, err := config.LoadConfig("../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	kb, err := krab.NewKrab(krab.Opts{Passphrase: cfg.IPFS.KrabPassword, DSPath: cfg.IPFS.KeystorePath})
	if err != nil {
		t.Fatal(err)
	}
	keystore, err := rtfs.NewKeystoreManager(kb)
	if err != nil {
		t.Fatal(err)
	}
	ipfs, err := rtfs.NewManager(cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port, keystore, time.Minute*5)
	if err != nil {
		t.Fatal(err)
	}
	api, err := new(cfg, gin.New(), ipfs, keystore, true, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	api.setupRoutes()
}
