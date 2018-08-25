package queue

import (
	"encoding/json"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// ProcessIPFSClusterPins is used to process messages sent to rabbitmq requesting be pinned to our cluster
// TODO: add in email notification and metric strategies
func (qm *QueueManager) ProcessIPFSClusterPins(msgs <-chan amqp.Delivery, cfg *config.TemporalConfig, db *gorm.DB) error {
	clusterManager, err := rtfs_cluster.Initialize(cfg.IPFSCluster.APIConnection.Host, cfg.IPFSCluster.APIConnection.Port)
	if err != nil {
		return err
	}

	qm.Logger.WithFields(log.Fields{
		"service": qm.QueueName,
	}).Info("processing ipfs cluster pins")

	for d := range msgs {

		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("new message detected")

		clusterAdd := IPFSClusterPin{}
		err = json.Unmarshal(d.Body, &clusterAdd)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   err.Error(),
			}).Error("error unmarshaling message")
			d.Ack(false)
			continue
		}

		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("successfully unmarshaled message, decoding hash string")

		encodedCid, err := clusterManager.DecodeHashString(clusterAdd.CID)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   err.Error(),
			}).Error("failed to decode hash string")
			d.Ack(false)
			continue
		}

		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Infof("pinning %s to cluster", clusterAdd.CID)

		err = clusterManager.Pin(encodedCid)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   err.Error(),
			}).Errorf("failed to pin %s to cluster", clusterAdd.CID)
			d.Ack(false)
			continue
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Infof("successfully pinned %s to cluster", clusterAdd.CID)
		d.Ack(false)
	}
	return nil
}
