package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/RTradeLtd/Temporal/bindings"
	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type PinPaymentConfirmation struct {
	TxHash        string `json:"tx_hash"`
	EthAddress    string `json:"eth_address"`
	UserName      string `json:"user_name"`
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
func (qm *QueueManager) ProcessPinPaymentConfirmation(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	paymentContractAddress := cfg.Ethereum.Contracts.PaymentContractAddress
	client, err := ethclient.Dial(cfg.Ethereum.Connection.INFURA.URL)
	if err != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"error":   err.Error(),
		}).Error("failed to connect to ethereum blockchain")
		return err
	}
	contract, err := bindings.NewPayments(common.HexToAddress(paymentContractAddress), client)
	if err != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"error":   err.Error(),
		}).Error("failed to generate payment contract handler")
		return err
	}
	qmEmail, err := Initialize(EmailSendQueue, cfg.RabbitMQ.URL, true, false)
	if err != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"error":   err.Error(),
		}).Error("failed to initialize connection to email send queue")
		return err
	}
	qmIpfs, err := Initialize(IpfsPinQueue, cfg.RabbitMQ.URL, true, false)
	if err != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"error":   err.Error(),
		}).Error("failed to initialize connection to ipfs pin queue")
		return err
	}
	paymentManager := models.NewPaymentManager(db)
	qm.Logger.WithFields(log.Fields{
		"service": qm.QueueName,
	}).Info("processing pin payment confirmations")
	for d := range msgs {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("new message detected")
		ppc := &PinPaymentConfirmation{}
		err = json.Unmarshal(d.Body, ppc)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   err.Error(),
			}).Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}
		tx, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(ppc.TxHash))
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service":     qm.QueueName,
				"eth_address": ppc.EthAddress,
				"tx_hash":     ppc.TxHash,
				"error":       err.Error(),
			}).Error("failed to get transaction hash")
			d.Ack(false)
			continue
		}
		if isPending {
			_, err := bind.WaitMined(context.Background(), client, tx)
			if err != nil {
				qm.Logger.WithFields(log.Fields{
					"service":     qm.QueueName,
					"eth_address": ppc.EthAddress,
					"tx_hash":     ppc.TxHash,
					"error":       err.Error(),
				}).Error("failed to wait for transaction to be mined")
				d.Ack(false)
				continue
			}
		}
		numberBig, valid := new(big.Int).SetString(ppc.PaymentNumber, 10)
		if !valid {
			addresses := []string{}
			addresses = append(addresses, ppc.EthAddress)
			es := EmailSend{
				Subject:     PaymentConfirmationFailedSubject,
				Content:     fmt.Sprintf(PaymentConfirmationFailedContent, ppc.ContentHash, "unable to convert string to big int"),
				ContentType: "",
				UserNames:   addresses,
			}
			err = qmEmail.PublishMessage(es)
			if err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"error":   err.Error(),
				}).Error("failed to publish message to email send queue")
			}
			qm.Logger.WithFields(log.Fields{
				"service":        qm.QueueName,
				"eth_address":    ppc.EthAddress,
				"payment_number": ppc.PaymentNumber,
				"error":          err.Error(),
			}).Error("failed to convert paymnet number to big int")
			d.Ack(false)
			continue
		}
		payment, err := contract.Payments(nil, common.HexToAddress(ppc.EthAddress), numberBig)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service":        qm.QueueName,
				"eth_address":    ppc.EthAddress,
				"payment_number": ppc.PaymentNumber,
				"error":          err.Error(),
			}).Error("failed to retrieve payment information from contract")
			d.Ack(false)
			continue
		}
		fmt.Printf("Payment struct \n%+v\n", payment)
		// now lets verify that the payment was indeed processed
		if payment.State != uint8(1) {
			addresses := []string{}
			addresses = append(addresses, ppc.EthAddress)
			es := EmailSend{
				Subject:     PaymentConfirmationFailedSubject,
				Content:     "payment unable to be processed, likely due to transaction failure or other contract runtime issue",
				ContentType: "",
				UserNames:   addresses,
			}
			err = qmEmail.PublishMessage(es)
			if err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"error":   err.Error(),
				}).Error("failed to publish message to email send queue")
			}
			qm.Logger.WithFields(log.Fields{
				"service":        qm.QueueName,
				"eth_address":    ppc.EthAddress,
				"payment_number": ppc.PaymentNumber,
				"error":          "unspecified transaction error",
			}).Error("transaction was mined, but contract code execution failed")
			d.Ack(false)
			continue
		}
		paymentFromDatabase, err := paymentManager.FindPaymentByNumberAndAddress(ppc.PaymentNumber, ppc.EthAddress)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service":     qm.QueueName,
				"eth_address": ppc.EthAddress,
				"error":       err.Error(),
			}).Error("failed to find payment in database")
			d.Ack(false)
			continue
		}
		// decide whether or not this should be handled here, or injected into the pin queue...
		// probably injected into the pin queue
		ip := IPFSPin{
			CID:              ppc.ContentHash,
			NetworkName:      paymentFromDatabase.NetworkName,
			UserName:         ppc.UserName,
			HoldTimeInMonths: paymentFromDatabase.HoldTimeInMonths,
		}

		err = qmIpfs.PublishMessageWithExchange(ip, PinExchange)
		if err != nil {
			addresses := []string{}
			addresses = append(addresses, ppc.EthAddress)
			es := EmailSend{
				Subject:     fmt.Sprintf("Critical Error: Unable to process IPFS Pin confirmation for content hash %s", ppc.ContentHash),
				Content:     "Please contact us at admin@rtradetechnologies.com and we will resolve this",
				ContentType: "",
				UserNames:   addresses,
			}
			errOne := qmEmail.PublishMessage(es)
			if errOne != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"error":   errOne.Error(),
				}).Error("failed to publish message to email send queue")
			}
			qm.Logger.WithFields(log.Fields{
				"service":     qm.QueueName,
				"eth_address": ppc.EthAddress,
				"error":       err.Error(),
			}).Error("critical error, failed to publish ipfs pin request for payment")
			d.Ack(false)
			continue
		}
		qm.Logger.WithFields(log.Fields{
			"service":        qm.QueueName,
			"eth_address":    ppc.EthAddress,
			"payment_number": ppc.PaymentNumber,
		}).Info("payment successfully processed")
		d.Ack(false)
	}
	return nil
}

// ProcessPinPaymentSubmissions is used to submit payments on behalf of a user. This does require them giving us the private key.
// while functional, this route isn't recommended as there are security risks involved. This will be upgraded over time so we can try
// to implement a more secure method. However keep in mind, this will always be "insecure". We may transition
// to letting the user sign the transactino, and we can broadcast the signed transaction
func (qm *QueueManager) ProcessPinPaymentSubmissions(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	paymentContractAddress := cfg.Ethereum.Contracts.PaymentContractAddress
	client, err := ethclient.Dial(cfg.Ethereum.Connection.INFURA.URL)
	if err != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"error":   err.Error(),
		}).Error("failed to connect to ethereum network")
		return err
	}
	contract, err := bindings.NewPayments(common.HexToAddress(paymentContractAddress), client)
	if err != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"error":   err.Error(),
		}).Error("failed to connect to connect to payment contract")
		return err
	}
	ppm := models.NewPaymentManager(db)
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"error":   err.Error(),
		}).Error("failed to connect to ipfs")
		return err
	}
	qm.Logger.WithFields(log.Fields{
		"service": qm.QueueName,
	}).Info("processing pin payment submissions")
	for d := range msgs {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("detected new message")
		pps := PinPaymentSubmission{}
		err = json.Unmarshal(d.Body, &pps)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   err.Error(),
			}).Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}
		k := keystore.Key{}
		err = k.UnmarshalJSON(pps.PrivateKey)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   err.Error(),
			}).Error("failed to unmarshal private key")
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
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   "bad type conversion",
			}).Error("failed to convert string to big int")
			d.Ack(false)
			continue
		}
		amount, valid := new(big.Int).SetString(pps.ChargeAmount, 10)
		if !valid {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   "bad type conversion",
			}).Error("failed to convert string to big int")
			d.Ack(false)
			continue
		}
		auth.GasLimit = 275000
		tx, err := contract.MakePayment(auth, h, v, r, s, num, method, amount, prefixed)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   err.Error(),
			}).Error("failed to submit payment to contract")
			d.Ack(false)
			continue
		}
		fmt.Println("successfully sent payment transaction, waiting for it to be mined")
		_, err = bind.WaitMined(context.Background(), client, tx)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service":     qm.QueueName,
				"eth_address": auth.From,
				"tx_hash":     tx.Hash().String(),
				"error":       err.Error(),
			}).Error("failed to wait for transaction to be mined")
			d.Ack(false)
			continue
		}
		paymentStruct, err := contract.Payments(nil, auth.From, num)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service":        qm.QueueName,
				"eth_address":    auth.From,
				"payment_number": num.String(),
				"error":          err.Error(),
			}).Error("failed to get payment from contract")
			d.Ack(false)
			continue
		}
		if paymentStruct.State != 1 {
			qm.Logger.WithFields(log.Fields{
				"service":        qm.QueueName,
				"eth_address":    auth.From,
				"payment_number": num.String(),
				"tx_hash":        tx.Hash().String(),
				"error":          "unspecifeid payment failure",
			}).Error("transaction was mined but payment failed to be processed")
			d.Ack(false)
			continue
		}
		paymentFromDB, err := ppm.FindPaymentByNumberAndAddress(num.String(), auth.From.String())
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service":        qm.QueueName,
				"eth_address":    auth.From,
				"payment_number": num.String(),
				"error":          err.Error(),
			}).Error("failed to find payment in database")
			d.Ack(false)
			continue
		}
		contentHash := paymentFromDB.ObjectName
		err = manager.Pin(contentHash)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service":        qm.QueueName,
				"eth_address":    auth.From,
				"payment_number": num.String(),
				"error":          err.Error(),
			}).Error("failed to pin content to ipfs")
			d.Ack(false)
			continue
		}
		qm.Logger.WithFields(log.Fields{
			"service":        qm.QueueName,
			"eth_address":    auth.From,
			"payment_number": num.String(),
		}).Info("payment successfully processed and content pinned to ipfs")
		d.Ack(false)
	}
	return nil
}
