package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/RTradeLtd/Temporal/bindings/payments"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
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

func ProcessPinPaymentConfirmation(msgs <-chan amqp.Delivery, db *gorm.DB, ipcPath, paymentContractAddress string) error {
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
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		return err
	}
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
		fmt.Println("unmarshaled payment")
		fmt.Printf("%+v\n", ppc)
		tx, isPending, err := client.TransactionByHash(context.Background(), common.HexToHash(ppc.TxHash))
		if err != nil {
			//TODO handle
			fmt.Println(err)
			// could be temporary error, so lets not ack
			continue
		}
		if isPending {
			_, err := bind.WaitMined(context.Background(), client, tx)
			if err != nil {
				//TODO handle
				fmt.Println(err)
				// could be a temporary error so lets not ack
				continue
			}
		}
		numberBig, valid := new(big.Int).SetString(ppc.PaymentNumber, 10)
		if !valid {
			// the message was improperly formatted so its garbagio
			fmt.Println("unable to convert string to big int")
			d.Ack(false)
			continue
		}
		payment, err := contract.Payments(nil, common.HexToAddress(ppc.EthAddress), numberBig)
		if err != nil {
			// TODO handle
			fmt.Println(err)
			// could be a temporary issue, so lets not ack
			continue
		}
		// now lets verify that the payment was indeed processed
		if payment.State != uint8(1) {
			// this means the payment wasn't actually confirmed, could be transaction rejection, etc...
			// by getting to this step in the code, it means the transaction has been mined so we need to ack this failure
			fmt.Println("payment unable to be processed, likely due to transaction failure or other contract runtime issue")
			d.Ack(false)
			continue
		}
		// here we have confirmed payment went through, so we can upload the file to our system
		err = manager.Pin(ppc.ContentHash)
		if err != nil {
			// this could be temporary so we wont ack
			fmt.Println(err)
			continue
		}
	}
	return nil
}
