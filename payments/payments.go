package payments

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"strings"

	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/utils"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
)

type PaymentManager struct {
	Contract *payments.Payments
	Client   *ethclient.Client
	Auth     *bind.TransactOpts
	DB       *gorm.DB
}

func NewPaymentManager(useIPC bool, ethKey, ethPass string, db *gorm.DB) (*PaymentManager, error) {
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
	pm.DB = db
	return &pm, nil
}

// RegisterPaymentForUploader is used to register a payment for the given uploader
func (pm *PaymentManager) RegisterPaymentForUploader(uploaderAddress string, contentHash string, retentionPeriodInMonths *big.Int, chargeAmountInWei *big.Int, method uint8, mqConnectionURL string) (*types.Transaction, error) {
	if method > 1 || method < 0 {
		return nil, errors.New("invalid payment method. 0 = RTC, 1 = ETH")
	}
	// since the contract function defines a fixed length byte array, we will need to convert a byte slice before calling the function
	var b [32]byte
	// convert hash to byte slice
	data := []byte(contentHash)
	// runs the byte slice through keccak256
	hashedCIDByte := crypto.Keccak256(data)
	// convert to hash
	hashedCID := common.BytesToHash(hashedCIDByte)
	// convert byte slice to byte array
	copy(b[:], hashedCID.Bytes()[:32])
	pm.Auth.GasPrice = big.NewInt(int64(22000000000))
	fmt.Println("submitting payment register transaction to the blockchain")
	// call the register payments function, indicating that we are expecting a user to upload the particular file
	// we hash the content identifier hash before submitting to the blockchain such that we preserve users privacy
	// but allow them to audit the contracts and data stores themselves, by hashing their plaintext content identifier hashes
	tx, err := pm.Contract.RegisterPayment(pm.Auth, common.HexToAddress(uploaderAddress), b, retentionPeriodInMonths, chargeAmountInWei, method)
	if err != nil {
		return nil, err
	}
	fmt.Println("spawning wait for transaction mined processor in go routine")
	// we'll do this in a go routine so we can return the tx information
	go pm.WaitForAndProcessMinedTransaction(tx, uploaderAddress, contentHash, hashedCID)
	fmt.Println("returning tx data")
	return tx, nil
}

// WaitForAndProcessMinedTransaction is used to wait for a payment register tx to be mined,
// so we can then update the database with the new payment information. After which we wait
// for the payment to be received, upon which we will upload the data to ipfs.
func (pm *PaymentManager) WaitForAndProcessMinedTransaction(tx *types.Transaction, uploaderAddress, contentHash string, hashedCID [32]byte) {
	var mined bool
	fmt.Println("waiting for tx to be mined")
	// wait for the transaction to be mined so we can read the payment id for the user
	for {
		if mined {
			fmt.Println("tx was mined")
			break
		}
		// check for the receipt
		receipt, err := pm.Client.TransactionReceipt(context.TODO(), tx.Hash())
		// if the error is not nil, and it is not of type not found, return
		if err != nil && err != ethereum.NotFound {
			fmt.Println("error retrieving tx receipt ", err)
			return
		}
		// cumulative gas used is greater than 0 so it has been mined
		if receipt.CumulativeGasUsed > 0 {
			mined = true
		}
	}
	fmt.Println("retrieving num payments")
	// read the total number of payments, so we can grab the latest payment id
	numPayments, err := pm.Contract.NumPayments(nil, common.HexToAddress(uploaderAddress))
	if err != nil {
		fmt.Println("error retrieving num payments ", err)
		return
	}
	fmt.Println("retrieving payment id")
	// grab the payment ID corresponding to this payment
	paymentID, err := pm.Contract.PaymentIDs(nil, common.HexToAddress(uploaderAddress), numPayments)
	if err != nil {
		fmt.Println("error retrieving payment id ", err)
		return
	}

	payment := models.Payment{
		UploaderAddress: uploaderAddress,
		CID:             contentHash,
		HashedCID:       fmt.Sprintf("%s", hex.EncodeToString(hashedCID[:])),
		PaymentID:       fmt.Sprintf("%s", hex.EncodeToString(paymentID[:])),
		Paid:            false,
	}
	pm.DB.Create(&payment)
}
