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
				err := json.Unmarshal(d.Body, &dfa)
				if err != nil {
					qm.LogError(err, "failed to unmarshal message")
					d.Ack(false)
					return
				}
				qm.LogInfo("message successfully unmarshaled")
				_, err = uploadManager.FindUploadByHashAndNetwork(dfa.Hash, dfa.NetworkName)
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
				if err != nil && err == gorm.ErrRecordNotFound {
					if _, err = uploadManager.NewUpload(
						dfa.Hash, "file",
						opts,
					); err != nil {
						qm.LogError(err, "failed to create new upload in database")
						d.Ack(false)
						return
					}
				} else {
					// this isn't a new upload so we shall upload the database;
					if _, err = uploadManager.UpdateUpload(dfa.HoldTimeInMonths, dfa.UserName, dfa.Hash, dfa.NetworkName); err != nil {
						qm.LogError(err, "failed to update upload")
						d.Ack(false)
						return
					}
				}
				qm.LogInfo("database file add processed")
				d.Ack(false)
				return // we must return here in order to trigger the wg.Done() defer
			}(d)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
		}
	}
}
