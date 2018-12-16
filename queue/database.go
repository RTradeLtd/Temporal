package queue

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/RTradeLtd/database/models"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessDatabaseFileAdds is used to process database file add messages
// No credit handling is done, as this route is only called to update the database
func (qm *Manager) ProcessDatabaseFileAdds(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	uploadManager := models.NewUploadManager(qm.db)
	qm.l.Info("processing database file adds")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processDatabaseFileAdd(d, wg, uploadManager)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
			return nil
		}
	}
}

func (qm *Manager) processDatabaseFileAdd(d amqp.Delivery, wg *sync.WaitGroup, um *models.UploadManager) {
	defer wg.Done()
	qm.l.Info("new database file add detected")
	dfa := DatabaseFileAdd{}
	// unmarshal the message body into the dfa struct
	if err := json.Unmarshal(d.Body, &dfa); err != nil {
		qm.l.Errorw(
			"failed to unmarshal message",
			"error", err.Error())
		d.Ack(false)
		return
	}
	upload, err := um.FindUploadByHashAndNetwork(dfa.Hash, dfa.NetworkName)
	if err != nil && err != gorm.ErrRecordNotFound {
		qm.l.Errorw(
			"failed to check database for upload",
			"error", err.Error(),
			"user", dfa.UserName)
		d.Ack(false)
		return
	}
	opts := models.UploadOptions{
		NetworkName:      dfa.NetworkName,
		Username:         dfa.UserName,
		HoldTimeInMonths: dfa.HoldTimeInMonths,
		Encrypted:        false,
	}
	if upload == nil {
		_, err = um.NewUpload(dfa.Hash, "file", opts)
	} else {
		// we have seen this upload before, so lets update the database record
		_, err = um.UpdateUpload(dfa.HoldTimeInMonths, dfa.UserName, dfa.Hash, dfa.NetworkName)
	}
	if err != nil {
		qm.l.Errorw(
			"failed to process database file add",
			"error", err.Error(),
			"user", dfa.UserName)
	} else {
		qm.l.Infow(
			"successfully processed database file add",
			"user", dfa.UserName)
	}
	d.Ack(false)
	return // we must return here in order to trigger the wg.Done() defer
}
