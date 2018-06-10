package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessPaymentReceivedQueue is used to process payment received messages
func ProcessPaymentReceivedQueue(msgs <-chan amqp.Delivery, db *gorm.DB) {
	ipfsManager, err := rtfs.Initialize("")
	if err != nil {
		log.Fatal(err)
	}
	for d := range msgs {
		var nullTime time.Time
		var payment models.Payment
		pr := PaymentReceived{}
		fmt.Println("unmarshaling payment received data")
		err := json.Unmarshal(d.Body, &pr)
		if err != nil {
			fmt.Println("error unmarhsaling data", err)
			d.Ack(false)
			continue
		}
		fmt.Printf("%+v\n", pr)
		fmt.Println("data unmarshaled successfully")
		db.Where("payment_id = ?", pr.PaymentID).Last(&payment)
		if payment.CreatedAt == nullTime {
			fmt.Println("payment is not a valid payment")
			d.Ack(false)
			continue
		}
		if payment.Paid {
			fmt.Println("payment already paid for")
			d.Ack(false)
			continue
		}
		fmt.Println("updating database with payment received")
		payment.Paid = true
		db.Model(&payment).Updates(map[string]interface{}{"paid": true})
		fmt.Println("database updated successfully, pinning to node")
		go ipfsManager.Pin(payment.CID)
		d.Ack(false)
	}
}

// ProcessPaymentRegisterQueue is used to process msgs sent to the payment
// register queue
func ProcessPaymentRegisterQueue(msgs <-chan amqp.Delivery, db *gorm.DB) {
	for d := range msgs {
		var nullTime time.Time
		var payment models.Payment
		pr := PaymentRegister{}
		fmt.Println("unmarshaling payment registered data")
		err := json.Unmarshal(d.Body, &pr)
		if err != nil {
			fmt.Println("error unmarshaling data", err)
			d.Ack(false)
			continue
		}
		fmt.Println("data unmarshaled successfully")
		db.Where("payment_id = ?", pr.PaymentID).Find(&payment)
		fmt.Println(payment)
		if payment.CreatedAt != nullTime {
			fmt.Println("payment is already in the database")
			d.Ack(false)
			continue
		}
		payment.CID = pr.CID
		payment.HashedCID = pr.HashedCID
		payment.PaymentID = pr.PaymentID
		payment.Paid = false
		fmt.Println("creating payment in database")
		db.Create(&payment)
		fmt.Println("payment entry in database created")
		d.Ack(false)
	}
}
