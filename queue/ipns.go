package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/RTradeLtd/Temporal/models"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

func ProcessIPNSPublishRequests(msgs <-chan amqp.Delivery, db *gorm.DB) error {
	var ipnsUpdate IPNSUpdate
	var resolve bool
	var switchErr bool
	rtfs, err := rtfs.Initialize("", "")
	um := models.NewUserManager(db)
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
		contentHash := ipnsUpdate.CID
		ttl, err := time.ParseDuration(ipnsUpdate.TTL)
		if err != nil {
			fmt.Println("error parsing ttl ", err)
			//TODO: handle
			d.Ack(false)
			continue
		}
		keyName := ipnsUpdate.Key
		lifetime, err := time.ParseDuration(ipnsUpdate.LifeTime)
		if err != nil {
			fmt.Println("error parsing lifetime ", err)
			//TODO: handle
			d.Ack(false)
			continue
		}
		resolveStr := ipnsUpdate.Resolve
		ethAddress := ipnsUpdate.EthAddress
		switch resolveStr {
		case "true":
			resolve = true
		case "false":
			resolve = false
		default:
			// TODO: handle
			fmt.Println("errror, resolve is neither \"true\" or \"false\" ")
			switchErr = true
		}
		if switchErr {
			// TODO: handle
			fmt.Println("errror, resolve is neither \"true\" or \"false\" ")
			d.Ack(false)
			continue
		}
		keyID, err := um.GetKeyIDByName(ethAddress, keyName)
		if err != nil {
			//TODO: handle
			fmt.Println("error", err.Error())
			d.Ack(false)
			continue
		}
		resp, err := rtfs.PublishToIPNSDetails(contentHash, keyID, lifetime, ttl, resolve)
		if err != nil {
			// TODO: handle
			fmt.Println("error publishing to ipns ", err)
			d.Ack(false)
			continue
		}
		fmt.Println("record published")
		fmt.Printf("name: %s\nvalue: %s\n", resp.Name, resp.Value)
		d.Ack(false)
	}
	return nil
}
