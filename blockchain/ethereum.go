package blockchain

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EthereumConnectionManager is our connect to the ethereum blockchain
type EthereumConnectionManager struct {
	Client *ethclient.Client
	Auth   *bind.TransactOpts
}

// GenerateEthereumConnectionManager generates our connection to the etheruem blockchain
func GenerateEthereumConnectionManager(cfg *config.TemporalConfig, connType string) (*EthereumConnectionManager, error) {
	var client *ethclient.Client
	var err error
	switch connType {
	case "infura":
		client, err = ethclient.Dial(cfg.Ethereum.Connection.INFURA.URL)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid connection type")
	}
	fileBytes, err := ioutil.ReadFile(cfg.Ethereum.Account.KeyFile)
	if err != nil {
		return nil, err
	}
	auth, err := bind.NewTransactor(strings.NewReader(string(fileBytes)), cfg.Ethereum.Account.KeyPass)
	if err != nil {
		return nil, err
	}
	return &EthereumConnectionManager{Client: client, Auth: auth}, nil
}
