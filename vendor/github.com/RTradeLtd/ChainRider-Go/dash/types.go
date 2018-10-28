package dash

import "net/http"

// Client is how we interact with the chain rider api
type Client struct {
	URL     string `json:"url"`
	Payload string `json:"payload"`
	Token   string `json:"token"`
	HC      *http.Client
}

// ConfigOpts are the configuration options for our session with teh api
type ConfigOpts struct {
	APIVersion      string `json:"api_version"`
	DigitalCurrency string `json:"digital_currency"`
	Blockchain      string `json:"blockchain"`
	Token           string `json:"token"`
}

const (
	payloadTemplate        = "{\n  \"token\": \"%s\"\n}"
	urlTemplate            = "https://api.chainrider.io/%s/%s/%s"
	rateLimitURL           = "https://api.chainrider.io/v1/ratelimit/"
	defaultAPIVersion      = "v1"
	defaultDigitalCurrency = "dash"
	defaultBlockchain      = "testnet"
)

// RateLimitResponse is the response given by the rate limit request
type RateLimitResponse struct {
	Message struct {
		Hour struct {
			Usage    int `json:"usage"`
			Limit    int `json:"limit"`
			TimeLeft int `json:"time_left"`
		} `json:"hour"`
		Day struct {
			Usage    int `json:"usage"`
			Limit    int `json:"limit"`
			TimeLeft int `json:"time_left"`
		} `json:"day"`
		Forward struct {
			Usage    int `json:"usage"`
			Limit    int `json:"limit"`
			TimeLeft int `json:"time_left"`
		} `json:"forward"`
	} `json:"message"`
}

// InformationResponse is the resposne from the information request
type InformationResponse struct {
	Info struct {
		Version         int     `json:"version"`
		InsightVerison  string  `json:"insightversion"`
		ProtocolVersion int     `json:"protocolversion"`
		Blocks          int     `json:"blocks"`
		TimeOffset      int     `json:"timeoffset"`
		Connections     int     `json:"connections"`
		Proxy           string  `json:"proxy"`
		Difficulty      float64 `json:"difficulty"`
		Testnet         bool    `json:"testnet"`
		RelayFee        float64 `json:"relayfee"`
		Errors          string  `json:"errors"`
		Network         string  `json:"network"`
	} `json:"info"`
}

// TransactionByHashResponse is a response for a given transaction hash request
type TransactionByHashResponse struct {
	TxID     string `json:"txid"`
	Version  int    `json:"version"`
	Locktime int    `json:"locktime"`
	Vin      []struct {
		TxID      string `json:"txid"`
		Vout      int    `json:"vout"`
		N         int    `json:"n"`
		ScriptSig struct {
			Hex string `json:"hex"`
			Asm string `json:"asm"`
		} `json:"scriptsig"`
		Addr            string  `json:"addr"`
		ValueSat        int     `json:"valueSat"`
		Value           float64 `json:"value"`
		DoubleSpentTxID string  `json:"doubleSpentTxID,omitempty"`
	} `json:"vin"`
	Vout []struct {
		Value        string `json:"value"` // this might need to be a string
		N            int    `json:"n"`
		ScriptPubKey struct {
			Hex       string   `json:"hex"`
			Asm       string   `json:"asm"`
			Addresses []string `json:"addresses"`
			Type      string   `json:"type"`
		} `json:"scriptPubKey"`
		SpentTxID   string `json:"spentTxId"`
		SpentIndex  int    `json:"spentIndex"`
		SpentHeight int    `json:"spentHeight"`
	} `json:"vout"`
	BlockHash     string  `json:"blockhash"`
	BlockHeight   int     `json:"blockheight"`
	Confirmations int     `json:"confirmations"`
	Time          int     `json:"time"`      // looks like its unix timestamp
	BlockTime     int     `json:"blocktime"` // looks like its unix timestamps
	ValueOut      float64 `json:"valueOut"`
	Size          int     `json:"size"`
	ValueIn       float64 `json:"valueIn"`
	Fees          float64 `json:"fees"`
	TxLock        bool    `json:"txlock"`
}

// TransactionsForAddressResponse is a collection of multiple transactions
type TransactionsForAddressResponse struct {
	PagesTotal   int                         `json:"pagesTotal"`
	Transactions []TransactionByHashResponse `json:"txs"`
}

// BlockByHashResponse is information fro a particular block
type BlockByHashResponse struct {
	Hash              string      `json:"hash"`
	Size              int         `json:"size"`
	Height            int         `json:"height"`
	Version           int         `json:"version"`
	MerkleRoot        string      `json:"merkleroot"`
	Tx                []string    `json:"tx"`
	Time              int         `json:"time"`
	Bits              string      `json:"bits"`
	ChainWork         string      `json:"chainwork"`
	Confirmations     int         `json:"confirmations"`
	PreviousBlockHash string      `json:"previousblockhash"`
	NextBlockHash     string      `json:"nextblockhash"`
	Reward            string      `json:"reward"`
	IsMainChain       bool        `json:"isMainChain"`
	PoolInfo          interface{} `json:"poolInfo"`
}

// LastBlockHashResponse is a resposne from the last block hash call
type LastBlockHashResponse struct {
	SyncTipHash   string `json:"syncTipHash"`
	LastBlockHash string `json:"lastblockhash"`
}

// BlockchainDataSyncStatusResponse is a response from the blockchain data sync call
type BlockchainDataSyncStatusResponse struct {
	Status           string `json:"status"`
	BlockchainHeight int    `json:"blockChainHeight"`
	SyncPercentage   int    `json:"syncPercentage"`
	Height           int    `json:"height"`
	Error            string `json:"error"`
	Type             string `json:"type"`
}

// CreatePaymentForwardResponse is a resposne from the create payment forward call
type CreatePaymentForwardResponse struct {
	PaymentForwardID     string  `json:"paymentforward_id"`
	PaymentAddress       string  `json:"payment_address"`
	DestinationAddress   string  `json:"destination_address"`
	CommissionFeePercent float64 `json:"commission_fee_percent"`
	MiningFeeDuffs       int     `json:"mining_fee_duffs"`
}

// GetPaymentForwardByIDResponse is a response a from get payment forward by id
type GetPaymentForwardByIDResponse struct {
	PaymentForwardID     string              `json:"paymentforward_id"`
	PaymentAddress       string              `json:"payment_address"`
	DestinationAddress   string              `json:"destination_address"`
	CommissionAddress    string              `json:"commission_address"`
	CommissionFeePercent float64             `json:"commission_fee_percent"`
	CreatedDate          string              `json:"created_date"`
	CallbackURL          string              `json:"callback_url"`
	MiningFeeDuffs       int                 `json:"mining_fee_duffs"`
	ProcessedTxs         []ProcessedTxObject `json:"processed_txs"`
}

// ProcessedTxObject is a transaction that has been processed for a payment forward
type ProcessedTxObject struct {
	InputTransactionHash string `json:"input_transaction_hash"`
	ReceivedAmountDuffs  int    `json:"received_amount_duffs"`
	TransactionHash      string `json:"transaction_hash"`
	// time at which the payment was forwarded
	ProcessedDate string `json:"processed_date"`
}
