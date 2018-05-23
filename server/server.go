package server

import (
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

var filesAddress common.Address
var usersAddress common.Address
var paymentsAddress common.Address

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
	// get the eth key (json file format)
	key := os.Getenv("ETH_KEY")
	// get the password to decrypt the key
	pass := os.Getenv("ETH_PASS")
	// url to connect t ofor the backend
	connURL := os.Getenv("CONN_URL")
	if key == "" || pass == "" || connURL == "" {
		log.Fatal("invalid key , password, or connection url")
	}
	// connect  to the network
	err := manager.ConnectToNetwork(connURL)
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
