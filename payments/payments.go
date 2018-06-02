package payments

import (
	"io/ioutil"
	"strings"

	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"
)

type PaymentManager struct {
	Contract *payments.Payments
	Client   *ethclient.Client
	Auth     *bind.TransactOpts
}

func NewPaymentManager(useIPC bool, ethKey, ethPass string) (*PaymentManager, error) {
	var pm PaymentManager
	var client *ethclient.Client
	file, err := ioutil.ReadFile(ethKey)
	if err != nil {
		return nil, err
	}
	switch useIPC {
	case true:
		client, err = ethclient.Dial(utils.IpcPath)
		if err != nil {
			return nil, err
		}
	case false:
		client, err = ethclient.Dial(utils.ConnectionURL)
		if err != nil {
			return nil, err
		}
	}
	auth, err := bind.NewTransactor(strings.NewReader(string(file)), ethPass)
	if err != nil {
		return nil, err
	}
	contract, err := payments.NewPayments(utils.PaymentsAddress, client)
	if err != nil {
		return nil, err
	}
	pm.Contract = contract
	pm.Client = client
	pm.Auth = auth
	return &pm, nil
}
