package queue

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/RTradeLtd/Temporal/rtfscluster"
	"github.com/RTradeLtd/database/models"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessIPFSClusterPins is used to process messages sent to rabbitmq requesting be pinned to our cluster
func (qm *Manager) ProcessIPFSClusterPins(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	clusterManager, err := rtfscluster.Initialize(qm.cfg.IPFSCluster.APIConnection.Host, qm.cfg.IPFSCluster.APIConnection.Port)
	if err != nil {
		return err
	}
	uploadManager := models.NewUploadManager(qm.db)
	qm.LogInfo("processing ipfs cluster pins")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go func(d amqp.Delivery) {
				defer wg.Done()
				qm.LogInfo("new message detected")
				clusterAdd := IPFSClusterPin{}
				if err = json.Unmarshal(d.Body, &clusterAdd); err != nil {
					qm.LogError(err, "failed to unmarshal message")
					d.Ack(false)
					return
				}
				if clusterAdd.NetworkName != "public" {
					qm.refundCredits(clusterAdd.UserName, "pin", clusterAdd.CreditCost)
					qm.LogError(err, "private networks not supported for ipfs cluster")
					d.Ack(false)
					return
				}
				qm.LogInfo("successfully unmarshaled message, decoding hash string")
				encodedCid, err := clusterManager.DecodeHashString(clusterAdd.CID)
				if err != nil {
					qm.refundCredits(clusterAdd.UserName, "pin", clusterAdd.CreditCost)
					qm.LogError(err, "failed to decode hash string")
					d.Ack(false)
					return
				}
				qm.LogInfo("pinning hash to cluster")
				if err = clusterManager.Pin(encodedCid); err != nil {
					qm.refundCredits(clusterAdd.UserName, "pin", clusterAdd.CreditCost)
					qm.LogError(err, "failed to pin hash to cluster")
					d.Ack(false)
					return
				}
				upload, err := uploadManager.FindUploadByHashAndNetwork(clusterAdd.CID, clusterAdd.NetworkName)
				if err != nil && err != gorm.ErrRecordNotFound {
					qm.LogError(err, "failed to search database for upload")
					d.Ack(false)
					return
				}
				if upload == nil {
					_, err = uploadManager.NewUpload(clusterAdd.CID, "pin-cluster", models.UploadOptions{
						NetworkName:      clusterAdd.NetworkName,
						Username:         clusterAdd.UserName,
						HoldTimeInMonths: clusterAdd.HoldTimeInMonths})
				} else {
					_, err = uploadManager.UpdateUpload(clusterAdd.HoldTimeInMonths, clusterAdd.UserName, clusterAdd.CID, clusterAdd.NetworkName)
				}
				if err != nil {
					qm.LogError(err, "failed to pin update database, but cluster was pinned")
				} else {
					qm.LogInfo("successfully pinned hash to cluster and updated database")
				}
				d.Ack(false)
				return // we must return here in order to trigger the wg.Done() defer
			}(d)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
			return nil
		}
	}
}
