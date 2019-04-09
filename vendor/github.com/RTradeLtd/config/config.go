package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

// LoadConfig loads a TemporalConfig from given filepath
func LoadConfig(configPath string) (*TemporalConfig, error) {
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var tCfg TemporalConfig
	if err = json.Unmarshal(raw, &tCfg); err != nil {
		return nil, err
	}

	tCfg.setDefaults()

	return &tCfg, nil
}

// GenerateConfig writes a empty TemporalConfig template to given filepath
func GenerateConfig(configPath string) error {
	template := &TemporalConfig{}
	template.setDefaults()
	b, err := json.Marshal(template)
	if err != nil {
		return err
	}

	var pretty bytes.Buffer
	if err = json.Indent(&pretty, b, "", "\t"); err != nil {
		return err
	}
	return ioutil.WriteFile(configPath, pretty.Bytes(), os.ModePerm)
}

func (t *TemporalConfig) setDefaults() {
	if t.LogDir == "" {
		t.LogDir = "/var/log/temporal/"
	}
	if len(t.API.Connection.CORS.AllowedOrigins) == 0 {
		t.API.Connection.CORS.AllowedOrigins = []string{"temporal.cloud", "backup.temporal.cloud"}
	}
}
