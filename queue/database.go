package queue

import (
	"encoding/json"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessDatabaseFileAdds is used to process database file add messages
func (qm *QueueManager) ProcessDatabaseFileAdds(msgs <-chan amqp.Delivery, db *gorm.DB) error {
	uploadManager := models.NewUploadManager(db)
	qm.LogInfo("processing database file adds")
	for d := range msgs {
		qm.LogInfo("detected new message")
		dfa := DatabaseFileAdd{}
		// unmarshal the message body into the dfa struct
		err := json.Unmarshal(d.Body, &dfa)
		if err != nil {
			qm.LogError(err, "failed to unmarshal message")
			d.Ack(false)
			continue
		}
		qm.LogInfo("message successfully unmarshaled")
		_, err = uploadManager.FindUploadByHashAndNetwork(dfa.Hash, dfa.NetworkName)
		if err != nil && err != gorm.ErrRecordNotFound {
			qm.LogError(err, "database check for upload failed")
			d.Ack(false)
			continue
		}
		opts := models.UploadOptions{
			NetworkName:      dfa.NetworkName,
			Username:         dfa.UserName,
			HoldTimeInMonths: dfa.HoldTimeInMonths,
			Encrypted:        false,
		}
		if err != nil && err == gorm.ErrRecordNotFound {
			if _, err = uploadManager.NewUpload(
				dfa.Hash, "file",
				models.UploadOptions{
					NetworkName:      dfa.NetworkName,
					Username:         dfa.UserName,
					HoldTimeInMonths: dfa.HoldTimeInMonths,
				},
			); err != nil {
				qm.LogError(err, "failed to create new upload in database")
				d.Ack(false)
				continue
			}
		} else {
			// this isn't a new upload so we shall upload the database;
			if _, err = uploadManager.UpdateUpload(dfa.HoldTimeInMonths, dfa.UserName, dfa.Hash, dfa.NetworkName); err != nil {
				qm.LogError(err, "failed to update upload")
				d.Ack(false)
				continue
			}
		}
		qm.LogInfo("database file add processed")
		d.Ack(false)
	}
	return nil
}
