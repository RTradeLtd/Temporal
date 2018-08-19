package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

type PinPaymentConfirmation struct {
	TxHash        string `json:"tx_hash"`
	EthAddress    string `json:"eth_address"`
	PaymentNumber string `json:"payment_number"`
	ContentHash   string `json:"content_hash"`
}

type PinPaymentSubmission struct {
	PrivateKey   []byte `json:"private_key"`
	Method       uint8  `json:"method"`
	Number       string `json:"number"`
	ChargeAmount string `json:"charge_amount"`
	// EthAddress string.... this is derived from the ethkey
	ContentHash string   `json:"content_hash"`
	H           [32]byte `json:"h"`
	V           uint8    `json:"v"`
	R           [32]byte `json:"r"`
	S           [32]byte `json:"s"`
	Hash        []byte   `json:"hash"`
	Sig         []byte   `json:"sig"`
	Prefixed    bool     `json:"prefixed"`
}

// ProcessPinPaymentConfirmation is used to process pin payment confirmations to inject content into TEMPORAL
// currently only supprots the private IPFS network
func ProcessPinPaymentConfirmation(msgs <-chan amqp.Delivery, db *gorm.DB, ipcPath, paymentContractAddress string, cfg *config.TemporalConfig) error {
	fmt.Println("dialing")
	client, err := ethclient.Dial(ipcPath)
	if err != nil {
		fmt.Println("error dialing", err)
		return err
	}
	fmt.Println("generating payment contract handler")
	contract, err := payments.NewPayments(common.HexToAddress(paymentContractAddress), client)
	if err != nil {
		fmt.Println("error generating payment contract", err)
		return err
	}
	qmEmail, err := Initialize(EmailSendQueue, cfg.RabbitMQ.URL, true)
	if err != nil {
		return err
	}
	qmIpfs, err := Initialize(IpfsPinQueue, cfg.RabbitMQ.URL, true)
	if err != nil {
		return err
	}
	paymentManager := models.NewPinPaymentManager(db)

	for d := range msgs {
		fmt.Println("payment detected")
		ppc := &PinPaymentConfirmation{}
		err = json.Unmarshal(d.Body, ppc)
		if err != nil {
			//TODO handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		tx, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(ppc.TxHash))
		if err != nil {
			fmt.Println(err)
			//TODO send email
			d.Ack(false)
			continue
		}
		if isPending {
			_, err := bind.WaitMined(context.Background(), client, tx)
			if err != nil {
				fmt.Println(err)
				//TODO send email
				d.Ack(false)
				continue
			}
		}
		numberBig, valid := new(big.Int).SetString(ppc.PaymentNumber, 10)
		if !valid {
			addresses := []string{}
			addresses = append(addresses, ppc.EthAddress)
			es := EmailSend{
				Subject:      PaymentConfirmationFailedSubject,
				Content:      fmt.Sprintf(PaymentConfirmationFailedContent, ppc.ContentHash, "unable to convert string to big int"),
				ContentType:  "",
				EthAddresses: addresses,
			}
			err = qmEmail.PublishMessage(es)
			if err != nil {
				fmt.Println("error publishing message ", err)
			}
			// the message was improperly formatted so its garbagio
			fmt.Println("unable to convert string to big int")
			d.Ack(false)
			continue
		}
		payment, err := contract.Payments(nil, common.HexToAddress(ppc.EthAddress), numberBig)
		if err != nil {
			fmt.Println(err)
			//TODO send email
			d.Ack(false)
			continue
		}
		fmt.Printf("Payment struct \n%+v\n", payment)
		// now lets verify that the payment was indeed processed
		if payment.State != uint8(1) {
			addresses := []string{}
			addresses = append(addresses, ppc.EthAddress)
			es := EmailSend{
				Subject:      PaymentConfirmationFailedSubject,
				Content:      "payment unable to be processed, likely due to transaction failure or other contract runtime issue",
				ContentType:  "",
				EthAddresses: addresses,
			}
			err = qmEmail.PublishMessage(es)
			if err != nil {
				fmt.Println("error publishing message ", err)
			}
			// this means the payment wasn't actually confirmed, could be transaction rejection, etc...
			// by getting to this step in the code, it means the transaction has been mined so we need to ack this failure
			fmt.Println("payment unable to be processed, likely due to transaction failure or other contract runtime issue")
			d.Ack(false)
			continue
		}
		paymentFromDatabase, err := paymentManager.RetrieveLatestPayment(ppc.EthAddress)
		if err != nil {
			//TODO: decide how we should handle
			fmt.Println("failed to retrieve latest payment ", err)
			d.Ack(false)
			continue
		}
		// decide whether or not this should be handled here, or injected into the pin queue...
		// probably injected into the pin queue
		ip := IPFSPin{
			CID:              ppc.ContentHash,
			NetworkName:      paymentFromDatabase.NetworkName,
			EthAddress:       ppc.EthAddress,
			HoldTimeInMonths: paymentFromDatabase.HoldTimeInMonths,
		}

		// DECIDE HOW WE SHOULD HANDLE FAILURES
		err = qmIpfs.PublishMessageWithExchange(ip, PinExchange)
		if err != nil {
			addresses := []string{}
			addresses = append(addresses, ppc.EthAddress)
			es := EmailSend{
				Subject:      fmt.Sprintf("Critical Error: Unable to process IPFS Pin confirmation for content hash %s", ppc.ContentHash),
				Content:      "Please contact us at admin@rtradetechnologies.com and we will resolve this",
				ContentType:  "",
				EthAddresses: addresses,
			}
			errOne := qmEmail.PublishMessage(es)
			if errOne != nil {
				fmt.Println("error publishing email to queue", errOne)
			}
			//TODO log and handle
			fmt.Println("error publishing pin to queue ", err)
			d.Ack(false)
			continue
		}
		d.Ack(false)
	}
	return nil
}

// ProcessPinPaymentSubmissions is used to submit payments on behalf of a user. This does require them giving us the private key.
// while functional, this route isn't recommended as there are security risks involved. This will be upgraded over time so we can try
// to implement a more secure method. However keep in mind, this will always be "insecure". We may transition
// to letting the user sign the transactino, and we can broadcast the signed transaction
func ProcessPinPaymentSubmissions(msgs <-chan amqp.Delivery, db *gorm.DB, ipcPath, paymentContractAddress string) error {
	client, err := ethclient.Dial(ipcPath)
	if err != nil {
		return err
	}
	contract, err := payments.NewPayments(common.HexToAddress(paymentContractAddress), client)
	if err != nil {
		return err
	}
	ppm := models.NewPinPaymentManager(db)
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		return err
	}
	for d := range msgs {
		fmt.Println("delivery detected")
		pps := PinPaymentSubmission{}
		err = json.Unmarshal(d.Body, &pps)
		if err != nil {
			fmt.Println("error unmarshaling", err)
			d.Ack(false)
			continue
		}
		k := keystore.Key{}
		err = k.UnmarshalJSON(pps.PrivateKey)
		if err != nil {
			fmt.Println("error unmarshaling private key", err)
			d.Ack(false)
			continue
		}
		auth := bind.NewKeyedTransactor(k.PrivateKey)
		h := pps.H
		v := pps.V
		r := pps.R
		s := pps.S
		method := pps.Method
		prefixed := pps.Prefixed
		num, valid := new(big.Int).SetString(pps.Number, 10)
		if !valid {
			fmt.Println("unable to convert payment number from string to big int")
			d.Ack(false)
			continue
		}
		amount, valid := new(big.Int).SetString(pps.ChargeAmount, 10)
		if !valid {
			fmt.Println("unable to convert charge amount from string to big int")
			d.Ack(false)
			continue
		}
		auth.GasLimit = 275000
		tx, err := contract.MakePayment(auth, h, v, r, s, num, method, amount, prefixed)
		if err != nil {
			fmt.Println("error making payment", err)
			d.Ack(false)
			continue
		}
		fmt.Println("successfully sent payment transaction, waiting for it to be mined")
		_, err = bind.WaitMined(context.Background(), client, tx)
		if err != nil {
			fmt.Println("error waiting for tx to be mined", err)
			d.Ack(false)
			continue
		}
		paymentStruct, err := contract.Payments(nil, auth.From, num)
		if err != nil {
			//TODO: add error handling (msg client via email notifying failure)
			fmt.Println("error retrieving payments", err)
			d.Ack(false)
			continue
		}
		if paymentStruct.State != 1 {
			fmt.Println("error occured while processing payment and the upload will not be processed")
			d.Ack(false)
			continue
		}
		paymentFromDB, err := ppm.FindPaymentByNumberAndAddress(num.String(), auth.From.String())
		if err != nil {
			fmt.Println("erorr reading payment from database", err)
			d.Ack(false)
			continue
		}
		contentHash := paymentFromDB.ContentHash
		err = manager.Pin(contentHash)
		if err != nil {
			fmt.Println("error pinning to IPFS", err)
			d.Ack(false)
			continue
		}
		fmt.Println("Content pinned to IPFS")
		d.Ack(false)
	}
	return nil
}
