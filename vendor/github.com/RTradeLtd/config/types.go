package config

// TemporalConfig defines Temporal configuration fields
type TemporalConfig struct {
	API         `json:"api,omitempty"`
	Database    `json:"database,omitempty"`
	IPFS        `json:"ipfs,omitempty"`
	IPFSCluster `json:"ipfs_cluster,omitempty"`
	MINIO       `json:"minio,omitempty"`
	RabbitMQ    `json:"rabbitmq,omitempty"`
	AWS         `json:"aws,omitempty"`
	Sendgrid    `json:"sendgrid,omitempty"`
	Ethereum    `json:"ethereum,omitempty"`
	Wallets     `json:"wallets,omitempty"`
	APIKeys     `json:"api_keys,omitempty"`
	Endpoints   `json:"endpoints,omitempty"`
}

// API configures the Temporal API
type API struct {
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
	JwtKey               string  `json:"jwt_key"`
	SizeLimitInGigaBytes string  `json:"size_limit_in_giga_bytes"`
	Payment              Payment `json:"payment"`
	LogFile              string  `json:"logfile"`
}

// Payment configures the GRPC Payment Server API
type Payment struct {
	Address  string `json:"address"`
	Port     string `json:"port"`
	Protocol string `json:"protocol"`
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

// AWS configures Temporal's connection to AWS
type AWS struct {
	KeyID  string `json:"key_id"`
	Secret string `json:"secret"`
}

// Sendgrid configures Temporal's connection to Sendgrid
type Sendgrid struct {
	APIKey       string `json:"api_key"`
	EmailAddress string `json:"email_address"`
	EmailName    string `json:"email_name"`
}

// Ethereum configures Temporal's connection, and interaction with the Ethereum blockchain
type Ethereum struct {
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
		RTCAddress             string `json:"rtc_address"`
		PaymentContractAddress string `json:"payment_contract_address"`
	} `json:"contracts"`
}

// Wallets are the addresses of RTrade Ltd's wallets
type Wallets struct {
	ETH  string `json:"eth"`
	RTC  string `json:"rtc"`
	XMR  string `json:"xmr"`
	DASH string `json:"dash"`
	BTC  string `json:"btc"`
	LTC  string `json:"ltc"`
}

// APIKeys are the various API keys we use
type APIKeys struct {
	ChainRider string `json:"chain_rider"`
}

// Endpoints are various endpoints we connect to
type Endpoints struct {
	MoneroRPC string `json:"monero_rpc"`
	LensGRPC  string `json:"lens_grpc"`
	MongoDB   struct {
		URL              string `json:"url"`
		DB               string `json:"db"`
		UploadCollection string `json:"uploads"`
	} `json:"mongodb"`
}
