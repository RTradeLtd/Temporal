package queue

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/RTradeLtd/Temporal/rtfscluster"
	"github.com/RTradeLtd/database/v2/models"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessIPFSClusterPins is used to process messages sent to rabbitmq requesting be pinned to our cluster
func (qm *Manager) ProcessIPFSClusterPins(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	clusterManager, err := rtfscluster.Initialize(ctx, qm.cfg.IPFSCluster.APIConnection.Host, qm.cfg.IPFSCluster.APIConnection.Port)
	if err != nil {
		return err
	}
	uploadManager := models.NewUploadManager(qm.db)
	qm.l.Info("processing ipfs cluster pin requests")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processIPFSClusterPin(ctx, d, wg, clusterManager, uploadManager)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
			return nil
		case msg := <-qm.ErrCh:
			qm.Close()
			wg.Done()
			qm.l.Errorw(
				"a protocol connection error stopping rabbitmq was received",
				"error", msg.Error())
			return errors.New(ErrReconnect)
		}
	}
}

func (qm *Manager) processIPFSClusterPin(ctx context.Context, d amqp.Delivery, wg *sync.WaitGroup, cm *rtfscluster.ClusterManager, um *models.UploadManager) {
	defer wg.Done()
	qm.l.Info("new cluster pin request detected")
	clusterAdd := IPFSClusterPin{}
	if err := json.Unmarshal(d.Body, &clusterAdd); err != nil {
		qm.l.Errorw(
			"failed to unmarshal message",
			"error", err.Error())
		d.Ack(false)
		return
	}
	if clusterAdd.NetworkName != "public" {
		qm.l.Errorw(
			"private clustered networks not yet supported",
			"error", errors.New("private network clusters not supported").Error(),
			"cid", clusterAdd.CID,
			"user", clusterAdd.UserName)
		d.Ack(false)
		return
	}
	encodedCid, err := cm.DecodeHashString(clusterAdd.CID)
	if err != nil {
		qm.refundCredits(clusterAdd.UserName, "pin", clusterAdd.CreditCost)
		models.NewUsageManager(qm.db).ReduceDataUsage(clusterAdd.UserName, uint64(clusterAdd.Size))
		qm.l.Errorw(
			"bad cid format detected",
			"error", err.Error(),
			"cid", clusterAdd.CID,
			"user", clusterAdd.UserName)
		d.Ack(false)
		return
	}
	qm.l.Infow(
		"pinning has to cluster",
		"cid", clusterAdd.CID,
		"user", clusterAdd.UserName)
	if err = cm.Pin(ctx, encodedCid); err != nil && err.Error() != "json: Unmarshal(nil) (200)" {
		_ = qm.refundCredits(clusterAdd.UserName, "pin", clusterAdd.CreditCost)
		_ = models.NewUsageManager(qm.db).ReduceDataUsage(clusterAdd.UserName, uint64(clusterAdd.Size))
		qm.l.Errorw(
			"failed to pin hash to cluster",
			"error", err.Error(),
			"cid", clusterAdd.CID,
			"user", clusterAdd.UserName)
		d.Ack(false)
		return
	}
	upload, err := um.FindUploadByHashAndUserAndNetwork(clusterAdd.UserName, clusterAdd.CID, clusterAdd.NetworkName)
	if err != nil && err != gorm.ErrRecordNotFound {
		qm.l.Errorw(
			"failed to check database for upload",
			"error", err.Error(),
			"cid", clusterAdd.CID,
			"user", clusterAdd.UserName)
		d.Ack(false)
		return
	}
	if upload == nil {
		_, err = um.NewUpload(clusterAdd.CID, "pin-cluster", models.UploadOptions{
			NetworkName:      clusterAdd.NetworkName,
			Username:         clusterAdd.UserName,
			HoldTimeInMonths: clusterAdd.HoldTimeInMonths,
			FileName:         clusterAdd.FileName,
			Size:             clusterAdd.Size})
	} else {
		_, err = um.UpdateUpload(clusterAdd.HoldTimeInMonths, clusterAdd.UserName, clusterAdd.CID, clusterAdd.NetworkName)
	}
	if err != nil {
		qm.l.Errorw(
			"failed to update database",
			"error", err.Error(),
			"cid", clusterAdd.CID,
			"user", clusterAdd.UserName)
	} else {
		qm.l.Infow(
			"successfully processed cluster pin request",
			"cid", clusterAdd.CID,
			"user", clusterAdd.UserName)
	}
	d.Ack(false)
}
