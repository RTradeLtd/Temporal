package payments

import (
	"errors"
	"io/ioutil"

	"github.com/RTradeLtd/Temporal/bindings"
	"github.com/RTradeLtd/Temporal/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// PaymentManager is our payment service
type PaymentManager struct {
	Client   *ethclient.Client
	Auth     *bind.TransactOpts
	Contract *bindings.Payments
}

// GeneratePaymentManager is used to generate our payment manager
func GeneratePaymentManager(cfg *config.TemporalConfig, connectionType string) (*PaymentManager, error) {
	pm := PaymentManager{}
	var client *ethclient.Client
	var err error
	switch connectionType {
	case "infura":
		client, err = ethclient.Dial(cfg.Ethereum.Connection.INFURA.URL)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported connection type, must be INFURA, IPC, RPC")
	}
	pm.Client = client
	err = pm.unlockAccount(cfg)
	if err != nil {
		return nil, err
	}
	contract, err := bindings.NewPayments(
		common.HexToAddress(cfg.Ethereum.Contracts.PaymentContractAddress),
		pm.Client)
	if err != nil {
		return nil, err
	}
	pm.Contract = contract
	return &pm, nil
}

func (pm *PaymentManager) unlockAccount(cfg *config.TemporalConfig) error {
	fileBytes, err := ioutil.ReadFile(
		cfg.Ethereum.Account.KeyFile)
	if err != nil {
		return err
	}
	pk, err := keystore.DecryptKey(
		fileBytes,
		cfg.Ethereum.Account.KeyPass)
	if err != nil {
		return err
	}
	auth := bind.NewKeyedTransactor(pk.PrivateKey)
	pm.Auth = auth
	return nil
}
