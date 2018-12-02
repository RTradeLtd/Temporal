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
func (qm *Manager) ProcessDatabaseFileAdds(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery, db *gorm.DB) error {
	uploadManager := models.NewUploadManager(db)
	qm.LogInfo("processing database file adds")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go func(d amqp.Delivery) {
				defer wg.Done()
				qm.LogInfo("detected new message")
				dfa := DatabaseFileAdd{}
				// unmarshal the message body into the dfa struct
				if err := json.Unmarshal(d.Body, &dfa); err != nil {
					qm.LogError(err, "failed to unmarshal message")
					d.Ack(false)
					return
				}
				qm.LogInfo("message successfully unmarshaled")
				upload, err := uploadManager.FindUploadByHashAndNetwork(dfa.Hash, dfa.NetworkName)
				if err != nil && err != gorm.ErrRecordNotFound {
					qm.LogError(err, "database check for upload failed")
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
					_, err = uploadManager.NewUpload(dfa.Hash, "file", opts)
				} else {
					// we have seen this upload before, so lets update teh database record
					_, err = uploadManager.UpdateUpload(dfa.HoldTimeInMonths, dfa.UserName, dfa.Hash, dfa.NetworkName)
				}
				if err != nil {
					qm.LogError(err, "failed to process database file upload")
				} else {
					qm.LogInfo("database file add processed")
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
