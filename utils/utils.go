package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// FilesAddress is the address of the files contract
var FilesAddress = common.HexToAddress("0x4863bc94E981AdcCA4627F56838079333f3D3700")

// UsersAddress is the address of the users contract
var UsersAddress = common.HexToAddress("0x1800fF6b7BFaa6223B90B1d791Bc6a8c582110CA")

// PaymentsAddress is the address of the payments contract
var PaymentsAddress = common.HexToAddress("0x3b2fD241378a326Af998E4243aA76fE8b8414dEe")

// ConnectionURL is the url used to connect to geth via rpc
var ConnectionURL = "http://127.0.0.1:8545"

// IpcPath is the file path used to connect to geth via ipc
var IpcPath = "/media/solidity/fuck/Rinkeby/datadir/geth.ipc"

// GenerateKeccak256HashFromString is  used to generate a keccak256 hash
// from string data into a format that is needed when making smart contract calls
func GenerateKeccak256HashFromString(data string) [32]byte {
	// this will hold the hashed data
	var b [32]byte
	// convert data into byte
	dataByte := []byte(data)
	// generate hash of the data
	hashedDataByte := crypto.Keccak256(dataByte)
	hash := common.BytesToHash(hashedDataByte)
	copy(b[:], hash.Bytes()[:32])
	return b
}
