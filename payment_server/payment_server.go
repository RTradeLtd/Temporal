package payment_server

import (
	"crypto/ecdsa"
	"io/ioutil"
	"math/big"
	"strings"

	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
	"github.com/onrik/ethrpc"
)

type PaymentManager struct {
	Contract   *payments.Payments
	Client     *ethclient.Client
	EthRPC     *ethrpc.EthRPC
	Auth       *bind.TransactOpts
	PrivateKey *ecdsa.PrivateKey
	DB         *gorm.DB
}

type SignedMessage struct {
	H []byte `json:"h"`
	R []byte `json:"r"`
	S []byte `json:"s"`
	V uint8  `json:"v"`
}

func GenerateSignedPaymentMessage(ethAddress common.Address, paymentMethod uint8, paymentNumber, chargeAmountInWei *big.Int) []byte {
	//  return keccak256(abi.encodePacked(msg.sender, _paymentNumber, _paymentMethod, _chargeAmountInWei));
	return SoliditySHA3(
		Address(ethAddress),
		Uint256(paymentNumber),
		Uint8(paymentMethod),
		Uint256(chargeAmountInWei),
	)
}

func (pm *PaymentManager) GenerateSignedPaymentMessage(ethAddress common.Address, paymentMethod uint8, paymentNumber, chargeAmountInWei *big.Int) (*SignedMessage, error) {
	//  return keccak256(abi.encodePacked(msg.sender, _paymentNumber, _paymentMethod, _chargeAmountInWei));
	hashToSign := SoliditySHA3(
		Address(ethAddress),
		Uint256(paymentNumber),
		Uint8(paymentMethod),
		Uint256(chargeAmountInWei),
	)
	sig, err := crypto.Sign(hashToSign, pm.PrivateKey)
	if err != nil {
		return nil, err
	}
	msg := &SignedMessage{
		H: hashToSign,
		R: sig[0:32],
		S: sig[32:64],
		V: uint8(sig[64]) + 1,
	}
	return msg, nil
}

func NewPaymentManager(useIPC bool, ethKey, ethPass string, db *gorm.DB) (*PaymentManager, error) {
	var pm PaymentManager
	var client *ethclient.Client
	// create a file handler from the key file path
	file, err := ioutil.ReadFile(ethKey)
	if err != nil {
		return nil, err
	}
	// check if they are using IPC or RPC, and create the appropriate connection
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
	// create our RPC client, make sure we have a proper connection
	rpcClient := ethrpc.NewEthRPC(utils.ConnectionURL)
	_, err = rpcClient.Web3ClientVersion()
	if err != nil {
		return nil, err
	}
	// decrypt our key file
	auth, err := bind.NewTransactor(strings.NewReader(string(file)), ethPass)
	if err != nil {
		return nil, err
	}
	// establish a connection with the payments contract
	contract, err := payments.NewPayments(utils.PaymentsAddress, client)
	if err != nil {
		return nil, err
	}
	pm.Contract = contract
	pm.Client = client
	pm.EthRPC = rpcClient
	pm.Auth = auth
	pm.DB = db
	return &pm, nil
}
