package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/RTradeLtd/ipd-config/ipdcfg"
)

// TemporalConfig is a helper struct holding
// our config values
type TemporalConfig struct {
	Database struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"database"`
	API struct {
		Admin struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"admin"`
		Connection struct {
			Certificates struct {
				CertPath string `json:"cert_path"`
				KeyPath  string `json:"key_path"`
			}
			ListenAddress string `json:"listen_address"`
		} `json:"connection"`
	} `json:"api"`
	Ethereum struct {
		Account struct {
			Address string `json:"address"`
			KeyFile string `json:"key_file"`
			KeyPass string `json:"key_pass"`
		} `json:"account"`
		Connection struct {
			RPC struct {
				IP   string `json:"ip"`
				Port string `json:"port"`
			} `json:"rpc"`
			IPC struct {
				Path string `json:"path"`
			} `json:"ipc"`
		} `json:"connection"`
	} `json:"ethereum"`
	RabbitMQ struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	} `json:"rabbitmq"`
}

func config(cfgCid string) {
	var tCfg TemporalConfig
	configManager := ipdcfg.Initialize("")
	config := configManager.LoadConfig(cfgCid)
	fmt.Println(config)
	err := json.Unmarshal(config, &tCfg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", tCfg)
}
