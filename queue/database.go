package queue

import (
	"encoding/json"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// ProcessDatabaseFileAdds is used to process database file add messages
func (qm *QueueManager) ProcessDatabaseFileAdds(msgs <-chan amqp.Delivery, db *gorm.DB) {
	uploadManager := models.NewUploadManager(db)
	qm.Logger.WithFields(log.Fields{
		"service": qm.QueueName,
	}).Info("processing database file adds")

	for d := range msgs {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("detected new message")

		dfa := DatabaseFileAdd{}
		// unmarshal the message body into the dfa struct
		err := json.Unmarshal(d.Body, &dfa)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   err.Error(),
			}).Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}

		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"user":    dfa.UserName,
		}).Info("message successfully unmarshaled")

		_, err = uploadManager.FindUploadByHashAndNetwork(dfa.Hash, dfa.NetworkName)
		if err != nil && err != gorm.ErrRecordNotFound {

			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"user":    dfa.UserName,
				"error":   err.Error(),
			}).Error("database check for upload failed")

			d.Ack(false)
			continue
		}
		if err != nil && err == gorm.ErrRecordNotFound {
			_, err = uploadManager.NewUpload(dfa.Hash, "file", dfa.NetworkName, dfa.UserName, dfa.HoldTimeInMonths)
			if err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"user":    dfa.UserName,
					"error":   err.Error(),
				}).Error("failed to create new upload in database")
			}
		} else {
			// this isn't a new upload so we shall upload the database
			_, err = uploadManager.UpdateUpload(dfa.HoldTimeInMonths, dfa.UserName, dfa.Hash, dfa.NetworkName)
			if err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"user":    dfa.UserName,
					"error":   err.Error(),
				}).Error("failed to update upload in database")
				d.Ack(false)
				continue
			}
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"user":    dfa.UserName,
		}).Infof("database file add for hash %s successfully processed", dfa.Hash)
		d.Ack(false)
	}
}
