package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func LoadConfig(configPath string) (*TemporalConfig, error) {
	var tCfg TemporalConfig
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw, &tCfg)
	if err != nil {
		return nil, err
	}
	return &tCfg, nil
}

func GenerateConfig(configPath string) error {
	template := &TemporalConfig{}
	b, err := json.Marshal(template)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(configPath, b, os.ModePerm)
}
