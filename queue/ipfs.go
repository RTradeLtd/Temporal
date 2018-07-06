package queue

import (
	"encoding/json"
	"fmt"

	"github.com/RTradeLtd/Temporal/rtfs"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessIpfsClusterQueue is used to process msgs sent to the ipfs cluster queue
func ProcessIpfsClusterQueue(msgs <-chan amqp.Delivery, db *gorm.DB) {
	var clusterPin IpfsClusterPin
	clusterManager := rtfs_cluster.Initialize()
	for d := range msgs {
		err := json.Unmarshal(d.Body, &clusterPin)
		if err != nil {
			fmt.Println("error unmarshaling data ", err)
			// TODO: handle error
			d.Ack(false)
			continue
		}
		contentHash := clusterPin.CID
		decodedContentHash, err := clusterManager.DecodeHashString(contentHash)
		if err != nil {
			fmt.Println("error decoded content hash to cid ", err)
			//TODO: handle error
			d.Ack(false)
			continue
		}
		err = clusterManager.Pin(decodedContentHash)
		if err != nil {
			fmt.Println("error pinning to cluster ", err)
			//TODO: handle error
			d.Ack(false)
			continue
		}
		fmt.Println("content pinned to cluster ", contentHash)
		d.Ack(false)

	}
}

// ProccessIPFSPins is used to process IPFS pin requests
func ProccessIPFSPins(msgs <-chan amqp.Delivery, db *gorm.DB) {
	userManager := models.NewUserManager(db)
	//uploadManager := models.NewUploadManager(db)
	networkManager := models.NewHostedIPFSNetworkManager(db)
	for d := range msgs {
		pin := &IPFSPin{}
		err := json.Unmarshal(d.Body, pin)
		if err != nil {
			//TODO log and handle
			fmt.Println(err)
			d.Ack(false)
		}
		apiURL := ""
		if pin.NetworkName != "public" {
			canAccess, err := userManager.CheckIfUserHasAccessToNetwork(pin.EthAddress, pin.NetworkName)
			if err != nil {
				//TODO log and handle
				fmt.Println(err)
				d.Ack(false)
				continue
			}
			if !canAccess {
				//TODO log 	and handle
				fmt.Println(err)
				fmt.Println("unauthorized access to private net ", pin.NetworkName)
				d.Ack(false)
				continue
			}
			url, err := networkManager.GetAPIURLByName(pin.NetworkName)
			if err != nil {
				//TODO log and handle
				fmt.Println(err)
				d.Ack(false)
			}
			apiURL = url
		}
		ipfsManager, err := rtfs.Initialize("", apiURL)
		if err != nil {
			//TODO log and handle
			// We aren't acknowledging this particular message
			// since it may be a temporary issue
			fmt.Println(err)
			continue
		}
		err = ipfsManager.Pin(pin.CID)
		if err != nil {
			//TODO log and handle
			// we aren't acknowlding this since it could be a temporary failure
			fmt.Println(err)
			fmt.Println("error pinning to network ", pin.NetworkName)
			continue
		}
	}
}

/* TODO: need to research a temporary storage backend to host the files temporarily
before uploading them to IPFS.
func ProcessIPFSFileAdds(msgs <-chan amqp.Delivery, db *gorm.DB) error {
	users := models.NewUserManager(db)
	uploads := models.NewUploadManager(db)
	networks := models.NewHostedIPFSNetworkManager(db)
}*/
