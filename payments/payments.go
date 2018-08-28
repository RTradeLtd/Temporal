package payments

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/RTradeLtd/Temporal/bindings"
	"github.com/RTradeLtd/Temporal/config"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
)

// PaymentService is our payment service
type PaymentService struct {
	Client   *ethclient.Client
	Auth     *bind.TransactOpts
	Contract *bindings.Payments
}

// GeneratePaymentManager is used to generate our payment manager
func GeneratePaymentManager(db *gorm.DB, cfg *config.TemporalConfig, connectionType string) (*PaymentService, error) {
	ps := PaymentService{}
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
	ps.Client = client
	err = ps.unlockAccount(cfg)
	if err != nil {
		return nil, err
	}
	contract, err := bindings.NewPayments(
		common.HexToAddress(cfg.Ethereum.Contracts.PaymentContractAddress),
		ps.Client)
	if err != nil {
		return nil, err
	}
	ps.Contract = contract
	return &ps, nil
}

func (ps *PaymentService) unlockAccount(cfg *config.TemporalConfig) error {
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
	ps.Auth = auth
	return nil
}

// ProcessPayments is used to process payments made to the smart contract
func (ps *PaymentService) ProcessPayments() error {
	var ch = make(chan *bindings.PaymentsPaymentMade)
	watchOpts := &bind.WatchOpts{Context: context.Background()}
	sub, err := ps.Contract.WatchPaymentMade(watchOpts, ch)
	if err != nil {
		return err
	}
	for {
		select {
		case err := <-sub.Err():
			fmt.Println("error parsing event", err.Error())
		case evLog := <-ch:
			fmt.Printf("%+v\n", evLog)
		}
	}
}
