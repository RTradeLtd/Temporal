package queue

import (
	"encoding/json"
	"fmt"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

func ProcessIPNSPublishRequests(msgs <-chan amqp.Delivery, db *gorm.DB) error {
	var ipnsUpdate IPNSUpdate
	rtfs, err := rtfs.Initialize("", "")
	if err != nil {
		return err
	}

	for d := range msgs {
		err := json.Unmarshal(d.Body, &ipnsUpdate)
		if err != nil {
			// TODO: handle
			fmt.Println("error unmarshaling into ipns update struct ", err)
			d.Ack(false)
			continue
		}
	}
	return nil
}
