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
		Password string `json:"password"`
	} `json:"database"`
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
	fmt.Println(tCfg)
}
