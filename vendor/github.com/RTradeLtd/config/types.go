package config

// TemporalConfig defines Temporal configuration fields
type TemporalConfig struct {
	API         `json:"api"`
	Database    `json:"database"`
	IPFS        `json:"ipfs"`
	IPFSCluster `json:"ipfs_cluster"`
	MINIO       `json:"minio"`
	RabbitMQ    `json:"rabbitmq"`
	AWS         struct {
		KeyID  string `json:"key_id"`
		Secret string `json:"secret"`
	} `json:"aws"`
	Sendgrid struct {
		APIKey       string `json:"api_key"`
		EmailAddress string `json:"email_address"`
		EmailName    string `json:"email_name"`
	} `json:"sendgrid"`
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
}

// API configures the Temporal API
type API struct {
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
}

// Database configures Temporal's connection to a Postgres database
type Database struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// IPFS configures Temporal's connection to an IPFS node
type IPFS struct {
	APIConnection struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"api_connection"`
}

// IPFSCluster configures Temporal's connection to an IPFS cluster
type IPFSCluster struct {
	APIConnection struct {
		Host string `json:"host"`
		Port string `json:"port"`
	} `json:"api_connection"`
}

// MINIO configures Temporal's connection to a Minio instance
type MINIO struct {
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	Connection struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	} `json:"connection"`
}

// RabbitMQ configures Temporal's connection to a RabbitMQ instance
type RabbitMQ struct {
	URL string `json:"url"`
}
