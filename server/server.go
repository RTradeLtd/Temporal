package server

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/RTradeLtd/Temporal/bindings/files"
	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/RTradeLtd/Temporal/bindings/users"
)

var filesAddress = common.HexToAddress("0x4863bc94E981AdcCA4627F56838079333f3D3700")
var usersAddress = common.HexToAddress("0x1800fF6b7BFaa6223B90B1d791Bc6a8c582110CA")
var paymentsAddress = common.HexToAddress("0x8Ed368f4e081cd55392e14EE6784FF5B32326247")
var connectionURL = "http://127.0.0.1:8545"

// ServerManager is a helper struct for interact with the server
type ServerManager struct {
	Client           *ethclient.Client
	Auth             *bind.TransactOpts
	PaymentsContract *payments.Payments
	UsersContract    *users.Users
	FilesContract    *files.Files
}

// Initialize is used to init the server manager
func Initialize() *ServerManager {
	// helper interface
	var manager ServerManager
	// get the password to decrypt the key
	keyFilePath := os.Getenv("KEY_FILE")

	pass := os.Getenv("ETH_PASS")
	if pass == "" {
		log.Fatal("ETH_PASS environment variable not set")
	}
	if keyFilePath == "" {
		log.Fatal("KEY_FILE environment variable not set")
	}
	file, err := ioutil.ReadFile(keyFilePath)
	if err != nil {
		log.Fatal(err)
	}
	key := string(file)
	// connect  to the network
	err = manager.ConnectToNetwork(connectionURL)
	if err != nil {
		log.Fatal(err)
	}
	// decrypt the key
	err = manager.Authenticate(key, pass)
	if err != nil {
		log.Fatal(err)
	}
	// initiate a connection to the files contract
	err = manager.NewFilesContract(filesAddress)
	if err != nil {
		log.Fatal(err)
	}
	// initiate a connection to the users contract
	err = manager.NewUsersContract(usersAddress)
	if err != nil {
		log.Fatal(err)
	}
	// initiate a connection to the payments contract
	err = manager.NewPaymentsContract(paymentsAddress)
	if err != nil {
		log.Fatal(err)
	}
	return &manager
}

// ConnectToNetwork is used to connect to the given evm network
func (sm *ServerManager) ConnectToNetwork(url string) error {
	// connectso the givne url, returning an ethclient object
	client, err := ethclient.Dial(url)
	if err != nil {
		return err
	}
	sm.Client = client
	return nil
}

// Authenticate is used to authenticate the ke
func (sm *ServerManager) Authenticate(key string, password string) error {
	// generats a *bind.TransactOpts object
	auth, err := bind.NewTransactor(strings.NewReader(key), password)
	if err != nil {
		return err
	}
	sm.Auth = auth
	return nil
}
