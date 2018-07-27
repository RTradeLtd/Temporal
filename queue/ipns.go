package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/models"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

type IPNSEntry struct {
	CID         string        `json:"cid"`
	LifeTime    time.Duration `json:"life_time"`
	TTL         time.Duration `json:"ttl"`
	Resolve     bool          `json:"resolve"`
	Key         string        `json:"key"`
	EthAddress  string        `json:"eth_address"`
	NetworkName string        `json:"network_name"`
}

// ProcessIPNSEntryCreationRequests is used to process IPNS entry creation requests
func ProcessIPNSEntryCreationRequests(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	ipfsManager, err := rtfs.Initialize("", "")
	err = ipfsManager.CreateKeystoreManager()
	if err != nil {
		return err
	}
	ipnsManager := models.NewIPNSManager(db)
	userManager := models.NewUserManager(db)
	networkManager := models.NewHostedIPFSNetworkManager(db)
	qmEmail, err := Initialize(IpnsEntryQueue, cfg.RabbitMQ.URL)
	if err != nil {
		return err
	}
	for d := range msgs {
		fmt.Println("ipns entry creation request detected")
		ie := IPNSEntry{}
		fmt.Println("unmarshaling response")
		err = json.Unmarshal(d.Body, &ie)
		if err != nil {
			//TODO log and handle
			fmt.Println("error unmarshaling ipns entry struct ", err)
			d.Ack(false)
			continue
		}
		fmt.Println("response unmarshaled")
		apiURL := ""
		if ie.NetworkName != "public" {
			fmt.Println("private ipfs network detected")
			canAccess, err := userManager.CheckIfUserHasAccessToNetwork(ie.EthAddress, ie.NetworkName)
			if err != nil {
				//TODO log and handle, decide how we should handle
				fmt.Println("error checking for private network acess ", err)
				d.Ack(false)
				continue
			}
			if !canAccess {
				addresses := []string{}
				addresses = append(addresses, ie.EthAddress)
				es := EmailSend{
					Subject:      IpfsPrivateNetworkUnauthorizedSubject,
					Content:      fmt.Sprintf("Unauthorized access to IPFS private network %s", ie.NetworkName),
					ContentType:  "",
					EthAddresses: addresses,
				}
				err = qmEmail.PublishMessage(es)
				if err != nil {
					//TODO log and handle
					fmt.Println("error publishing message ", err)
				}
				fmt.Println("unauthorized access to private net ", ie.NetworkName)
				d.Ack(false)
				continue
			}
			apiURLName, err := networkManager.GetAPIURLByName(ie.NetworkName)
			if err != nil {
				//TODO send email, log, handle
				fmt.Println("erro getting API url by name ", err)
				d.Ack(false)
				continue
			}
			apiURL = apiURLName
			ipfsManager, err = rtfs.Initialize("", apiURL)
			if err != nil {
				addresses := []string{}
				addresses = append(addresses, ie.EthAddress)
				es := EmailSend{
					Subject:      IpfsInitializationFailedSubject,
					Content:      fmt.Sprintf("Connection to IPFS failed due to the following error %s", err),
					ContentType:  "",
					EthAddresses: addresses,
				}
				errOne := qmEmail.PublishMessage(es)
				if errOne != nil {
					fmt.Println("error publishing message ", err)
				}
				fmt.Println(err)
				d.Ack(false)
				continue
			}
		}
		fmt.Println("publishing response")
		response, err := ipfsManager.PublishToIPNSDetails(ie.CID, ie.Key, ie.LifeTime, ie.TTL, ie.Resolve)
		if err != nil {
			fmt.Println("error publishing response")
			formattedContent := fmt.Sprintf(IpnsEntryFailedContent, ie.CID, ie.Key, err)
			addresses := []string{}
			addresses = append(addresses, ie.EthAddress)
			es := EmailSend{
				Subject:      IpnsEntryFailedSubject,
				Content:      formattedContent,
				ContentType:  "",
				EthAddresses: addresses,
			}
			errOne := qmEmail.PublishMessage(es)
			if errOne != nil {
				fmt.Println("error publishing message to email queue ", errOne)
			}
			fmt.Println("error publishing IPNS entry ", err)
			d.Ack(false)
			continue
		}
		_, err = ipnsManager.UpdateIPNSEntry(response.Name, ie.CID, ie.Key, ie.NetworkName, ie.LifeTime, ie.TTL)
		if err != nil {
			//TODO: decide how to handle
			fmt.Println("error adding IPNS entry to database ", err)
		}
		fmt.Println("response published successfully")
		fmt.Println("IPNS entry creation successful ", response)
		//TODO update database
		d.Ack(false)
	}
	return nil
}

// ProcessIPNSUpdates is used to process any IPNS updates, saving them to the database
func ProcessIPNSUpdates(msgs <-chan amqp.Delivery, db *gorm.DB) error {
	im := models.NewIPNSManager(db)
	//um := models.NewUserManager(db)
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		return err
	}
	err = manager.CreateKeystoreManager()
	if err != nil {
		return err
	}
	for d := range msgs {
		ipnsUpdate := IPNSUpdate{}
		err := json.Unmarshal(d.Body, &ipnsUpdate)
		if err != nil {
			//TODO handle
			fmt.Println("error unmarshaling into ipns update struct ", err)
			d.Ack(false)
			continue
		}
		ipnsHash := ipnsUpdate.IPNSHash
		ipfsHash := ipnsUpdate.CID
		key := ipnsUpdate.Key
		networkName := ipnsUpdate.NetworkName
		lifetime, err := time.ParseDuration(ipnsUpdate.LifeTime)
		if err != nil {
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		ttl, err := time.ParseDuration(ipnsUpdate.TTL)
		if err != nil {
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		_, err = im.UpdateIPNSEntry(ipnsHash, ipfsHash, key, networkName, lifetime, ttl)
		if err != nil {
			fmt.Println(err)
			d.Ack(false)
			continue
		}

	}
	return nil
}
