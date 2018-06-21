package rtfsp

import (
	"encoding/json"
	"os"

	config "github.com/ipfs/go-ipfs/repo/config"
)

func InitializeConfig(configFilePath string) (*config.Config, error) {
	var conf config.Config
	reader, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	err = json.NewDecoder(reader).Decode(&conf)
	return &conf, nil
}
