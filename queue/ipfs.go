package queue

import (
	"encoding/json"
	"fmt"

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
