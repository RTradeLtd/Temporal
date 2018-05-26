package main

import (
	"fmt"

	"github.com/RTradeLtd/ipd-config/ipdcfg"
)

// TemporalConfig is a helper struct holding
// our config values
type TemporalConfig struct {
	DatabasePassword       string `json:"database_password"`
	APIAdminUser           string `json:"api_admin_user"`
	APIAdminPass           string `json:"api_admin_pass"`
	APIJwtKeystring        string `json:"api_jwt_key"`
	APIListenAddress       string `json:"api_listen_address"`
	APICertificateCertPath string `json:"api_certificate_cert_path"`
	APICertificateKeyPath  string `json:"api_certificate_key_path"`
}

func config(cfgCid string) {
	var tCfg TemporalConfig
	configManager := ipdcfg.Initialize("")
	config := configManager.LoadConfig(cfgCid)
	for k, v := range config {
		switch k {
		case "database":
			tCfg.DatabasePassword = fmt.Sprint(v)
		}
	}
}
