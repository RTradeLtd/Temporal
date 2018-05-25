package server

import (
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// NewPaymentsContract is used to generate a new payment contract handler
func (sm *ServerManager) NewPaymentsContract(address common.Address) error {
	contract, err := payments.NewPayments(address, sm.Client)
	if err != nil {
		return err
	}
	sm.PaymentsContract = contract
	return nil
}

// RegisterPaymentForUploader is used to register a payment for the given uploader
func (sm *ServerManager) RegisterPaymentForUploader(uploaderAddress string, contentHash string, retentionPeriodInMonths *big.Int, chargeAmountInWei *big.Int, method uint8) (*types.Transaction, error) {
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
	// call the register payments function, indicating that we are expecting a user to upload the particular file
	// we hash the content identifier hash before submitting to the blockchain such that we preserve users privacy
	// but allow them to audit the contracts and data stores themselves, by hashing their plaintext content identifier hashes
	tx, err := sm.PaymentsContract.RegisterPayment(sm.Auth, common.HexToAddress(uploaderAddress), b, retentionPeriodInMonths, chargeAmountInWei, method)
	if err != nil {
		return nil, err
	}
	// in order to get their latest payment id, we need to get the total number of payments
	// since we just submitted the transaction, we will read off pending state
	numPayments, err := sm.PaymentsContract.NumPayments(&bind.CallOpts{Pending: true}, common.HexToAddress(uploaderAddress))
	if err != nil {
		return tx, err
	}
	// having their total number of payments, get the payment ID associated with their latest payment
	// again, isnce we just submitted the call, we will read off pending state
	paymentID, err := sm.PaymentsContract.PaymentIDs(&bind.CallOpts{Pending: true}, common.HexToAddress(uploaderAddress), numPayments)
	if err != nil {
		return tx, err
	}
	// construct a message to rabbitmq so we can save this paymetn information to the database
	pr := queue.PaymentRegister{
		UploaderAddress: uploaderAddress,
		CID:             contentHash,
		HashedCID:       hashedCID.String(),
		PaymentID:       string(paymentID[:]), // this converts the paymentID byte array to a string
	}
	// initialize a conenction to rabbitmq
	qm, err := queue.Initialize(queue.PaymentRegisterQueue)
	if err != nil {
		return tx, err
	}
	// publish the message to rabbitmq
	err = qm.PublishMessage(pr)
	if err != nil {
		return tx, err
	}
	return tx, nil
}

func (sm *ServerManager) WaitForAndProcessPaymentsReceivedEvent() {
	// create the channel for which we will receive payments on
	var ch = make(chan *payments.PaymentsPaymentReceivedNoIndex)
	// create a subscription for th eevent passing in messages to teh chanenl we just established
	sub, err := sm.PaymentsContract.WatchPaymentReceivedNoIndex(nil, ch)
	if err != nil {
		log.Fatal(err)
	}
	queueManager, err := queue.Initialize(queue.PaymentRegisterQueue)
	if err != nil {
		log.Fatal(err)
	}
	queueManager.PublishMessage("hello")
	// loop forever, waiting for and processing events
	for {
		select {
		case err := <-sub.Err():
			fmt.Println("Error parsing event ", err)
		case evLog := <-ch:
			uploader := evLog.Uploader
			paymentID := evLog.PaymentID
			chargeAmountInWei := evLog.Amount
			paymentMethod := evLog.Method
			pr := queue.PaymentRegister{}
			pr.UploaderAddress = uploader.String()
			fmt.Println(uploader, paymentID, chargeAmountInWei, paymentMethod)
		}
	}
}
