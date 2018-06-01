package server

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/RTradeLtd/Temporal/queue"
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
func (sm *ServerManager) RegisterPaymentForUploader(uploaderAddress string, contentHash string, retentionPeriodInMonths *big.Int, chargeAmountInWei *big.Int, method uint8, mqConnectionURL string) (*types.Transaction, error) {
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
	sm.Auth.GasPrice = big.NewInt(int64(22000000000))
	// call the register payments function, indicating that we are expecting a user to upload the particular file
	// we hash the content identifier hash before submitting to the blockchain such that we preserve users privacy
	// but allow them to audit the contracts and data stores themselves, by hashing their plaintext content identifier hashes
	tx, err := sm.PaymentsContract.RegisterPayment(sm.Auth, common.HexToAddress(uploaderAddress), b, retentionPeriodInMonths, chargeAmountInWei, method)
	if err != nil {
		return nil, err
	}
	go sm.RegisterWaitForAndProcessPaymentsRegisteredEventForAddress(uploaderAddress, contentHash, mqConnectionURL)

	return tx, nil
}

// RegisterWaitForAndProcessPaymentsRegisteredEventForAddress is used tolisten for payment register events from the payments contract
// for a particular user. It then sends a message to rabbit mq to update the database with the payment registration details. When
// we detect that a payment has been received following a payment registration (this is done via a seperate listener), we update the database
// as having received the payment, followed by uploading the particular content to ipfs
func (sm *ServerManager) RegisterWaitForAndProcessPaymentsRegisteredEventForAddress(address, cid, mqConnectionURL string) {
	var processed bool
	// this channel will receive events from the smart contract
	var ch = make(chan *payments.PaymentsPaymentRegistered)
	// create a subscription handle for this particular event
	sub, err := sm.PaymentsContract.WatchPaymentRegistered(nil, ch, []common.Address{common.HexToAddress(address)})
	if err != nil {
		log.Fatal(err)
	}
	// create a connection to rabbitmq
	queueManager, err := queue.Initialize(queue.PaymentRegisterQueue, mqConnectionURL)
	if err != nil {
		log.Fatal(err)
	}
	for {
		if processed {
			break
		}
		select {
		case err := <-sub.Err():
			fmt.Println("Error parsing event ", err)
			log.Fatal(err)
		case evLog := <-ch:
			pr := queue.PaymentRegister{}
			uploader := evLog.Uploader
			hashedCID := evLog.HashedCID
			paymentID := evLog.PaymentID
			pr.UploaderAddress = uploader.String()
			pr.CID = cid
			// the following is used to convert a byte-array into a byte-slice, followed by encoding to string
			pr.HashedCID = fmt.Sprintf("%s", hex.EncodeToString(hashedCID[:]))
			pr.PaymentID = fmt.Sprintf("0x%s", hex.EncodeToString(paymentID[:]))
			// publish the message to rabbitmq.
			// the rabbitmq worker will parse the event, and update the database
			queueManager.PublishMessage(pr)
			// set processed to true, allowing us to break out of the outer loop
			processed = true
			// break out of the inner loop
			break
		}
	}
}

// WaitForAndProcessPaymentsReceivedEvent is used to watch for for a payment received event. It
// then parses the data sending a message to rabbitmq. The rabbitmq worker will then mark the payment
// as having been received in hte database, followed by triggering an upload to ipfs
func (sm *ServerManager) WaitForAndProcessPaymentsReceivedEvent(mqConnectionURL string) {
	// create the channel for which we will receive payments on
	var ch = make(chan *payments.PaymentsPaymentReceivedNoIndex)
	// create a subscription for th eevent passing in messages to teh chanenl we just established
	sub, err := sm.PaymentsContract.WatchPaymentReceivedNoIndex(nil, ch)
	if err != nil {
		log.Fatal(err)
	}

	queueManager, err := queue.Initialize(queue.PaymentReceivedQueue, mqConnectionURL)
	if err != nil {
		log.Fatal(err)
	}
	// loop forever, waiting for and processing events
	for {
		select {
		case err := <-sub.Err():
			fmt.Println("Error parsing event ", err)
		case evLog := <-ch:
			uploader := evLog.Uploader
			paymentID := evLog.PaymentID
			pr := queue.PaymentReceived{}
			pr.UploaderAddress = uploader.String()
			pr.PaymentID = fmt.Sprintf("0x%s", hex.EncodeToString(paymentID[:]))
			queueManager.PublishMessage(pr)
		}
	}
}
