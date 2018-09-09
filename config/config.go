package config

import (
	"encoding/json"
	"io/ioutil"
)

// TemporalConfig is a helper struct holding
// our config values
type TemporalConfig struct {
	Database struct {
		Name     string `json:"name"`
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"database"`
	API struct {
		AdminUser  string `json:"admin_user"`
		Connection struct {
			Certificates struct {
				CertPath string `json:"cert_path"`
				KeyPath  string `json:"key_path"`
			}
			ListenAddress string `json:"listen_address"`
		} `json:"connection"`
		Sessions struct {
			AuthKey       string `json:"auth_key"`
			EncryptionKey string `json:"encryption_key"`
		} `json:"sessions"`
		RollbarToken         string `json:"rollbar_token"`
		JwtKey               string `json:"jwt_key"`
		SizeLimitInGigaBytes string `json:"size_limit_in_giga_bytes"`
	} `json:"api"`
	IPFS struct {
		APIConnection struct {
			Host string `json:"host"`
			Port string `json:"port"`
		} `json:"api_connection"`
	} `json:"ipfs"`
	IPFSCluster struct {
		APIConnection struct {
			Host string `json:"host"`
			Port string `json:"port"`
		} `json:"api_connection"`
	} `json:"ipfs_cluster"`
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
			INFURA struct {
				URL string `json:"url"`
			} `json:"infura"`
		} `json:"connection"`
		Contracts struct {
			PaymentContractAddress string `json:"payment_contract_address"`
		} `json:"contracts"`
	} `json:"ethereum"`
	RabbitMQ struct {
		URL string `json:"url"`
	} `json:"rabbitmq"`
	AWS struct {
		KeyID  string `json:"key_id"`
		Secret string `json:"secret"`
	} `json:"aws"`
	MINIO struct {
		AccessKey  string `json:"access_key"`
		SecretKey  string `json:"secret_key"`
		Connection struct {
			IP   string `json:"ip"`
			Port string `json:"port"`
		} `json:"connection"`
	} `json:"minio"`
	Sendgrid struct {
		APIKey       string `json:"api_key"`
		EmailAddress string `json:"email_address"`
		EmailName    string `json:"email_name"`
	} `json:"sendgrid"`
}

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

/*
// LoadConfig is used to load a config object, from the cid
func LoadConfig(cfgCid string) *TemporalConfig {
	var tCfg TemporalConfig
	configManager := ipdcfg.Initialize("")
	config := configManager.LoadConfig(cfgCid)
	err := json.Unmarshal(config, &tCfg)
	if err != nil {
		log.Fatal(err)
	}
	return &tCfg
}
*/
