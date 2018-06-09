package server

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/RTradeLtd/Temporal/bindings/files"
	"github.com/RTradeLtd/Temporal/bindings/users"
	"github.com/RTradeLtd/Temporal/utils"
)

// ServerManager is a helper struct for interact with the server
type ServerManager struct {
	Client        *ethclient.Client
	Auth          *bind.TransactOpts
	UsersContract *users.Users
	FilesContract *files.Files
}

// Initialize is used to init the server manager
func Initialize(useIPC bool, ethKey, ethPass string) *ServerManager {
	// helper interface
	var manager ServerManager
	// get the password to decrypt the key

	pass := ethPass
	if pass == "" {
		log.Fatal("ETH_PASS environment variable not set")
	}
	keyFilePath := ethKey

	if keyFilePath == "" {
		log.Fatal("KEY_FILE environment variable not set")
	}
	file, err := ioutil.ReadFile(keyFilePath)
	if err != nil {
		log.Fatal(err)
	}
	key := string(file)

	if useIPC {
		// connect  to the network
		err = manager.ConnectToNetwork(utils.IpcPath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = manager.ConnectToNetwork(utils.ConnectionURL)
		if err != nil {
			log.Fatal(err)
		}
	}
	// decrypt the key
	err = manager.Authenticate(key, pass)
	if err != nil {
		log.Fatal(err)
	}
	// initiate a connection to the files contract
	err = manager.NewFilesContract(utils.FilesAddress)
	if err != nil {
		log.Fatal(err)
	}
	// initiate a connection to the users contract
	err = manager.NewUsersContract(utils.UsersAddress)
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
