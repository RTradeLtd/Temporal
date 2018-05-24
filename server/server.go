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

var filesAddress = common.HexToAddress("0xb4452c00e62F8FE634AbCB8E1a8d7eC2aC42b796")
var usersAddress = common.HexToAddress("0xC283BfEf5Eeb6A88d51Ddcf26E59a5e1A22C0280")
var paymentsAddress = common.HexToAddress("0x492710A119dF8133aAdd72d0A1e37D63B5F2fdfA")
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
	file, err := ioutil.ReadFile("/tmp/key_file")
	if err != nil {
		log.Fatal("err")
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
