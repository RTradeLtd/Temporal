package queue

import (
	"encoding/json"
	"fmt"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessIPFSClusterPins is used to process messages sent to rabbitmq requesting be pinned to our cluster
// TODO: add in email notification and metric strategies
func (qm *QueueManager) ProcessIPFSClusterPins(msgs <-chan amqp.Delivery, cfg *config.TemporalConfig, db *gorm.DB) error {
	clusterManager, err := rtfs_cluster.Initialize(cfg.IPFSCluster.APIConnection.Host, cfg.IPFSCluster.APIConnection.Port)
	if err != nil {
		return err
	}
	for d := range msgs {
		clusterAdd := IPFSClusterPin{}
		err = json.Unmarshal(d.Body, &clusterAdd)
		if err != nil {
			fmt.Println("error unmarshaling cluster add message ", err)
			d.Ack(false)
			continue
		}
		if clusterAdd.NetworkName != "" {
			//TODO implement adds to private clusters
			d.Ack(false)
			continue
		}
		encodedCid, err := clusterManager.DecodeHashString(clusterAdd.CID)
		if err != nil {
			fmt.Println("failed to encode hash string to cid object ", err)
			d.Ack(false)
			continue
		}
		err = clusterManager.Pin(encodedCid)
		if err != nil {
			fmt.Println("failed to pin content to ipfs cluster ", err)
			d.Ack(false)
			continue
		}
		fmt.Println("successfully pinned content to ipfs cluster")
		d.Ack(false)
	}
	return nil
}
