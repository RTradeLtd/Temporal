package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

var nilTime time.Time

// ProcessDatabaseFileAdds is used to process database file add messages
func ProcessDatabaseFileAdds(msgs <-chan amqp.Delivery, db *gorm.DB) {
	uploadManager := models.NewUploadManager(db)
	for d := range msgs {
		fmt.Println("detected new message")
		dfa := DatabaseFileAdd{}
		// unmarshal the message body into the dfa struct
		err := json.Unmarshal(d.Body, &dfa)
		if err != nil {
			d.Ack(false)
			continue
		}
		fmt.Println("processing database update for content hash", dfa.Hash)
		_, err = uploadManager.FindUploadByHashAndNetwork(dfa.Hash, dfa.NetworkName)
		if err != nil && err != gorm.ErrRecordNotFound {
			fmt.Println("error looking for upload in database ", err)
			//TODO send email
			d.Ack(false)
			continue
		}
		if err != nil && err == gorm.ErrRecordNotFound {
			_, err = uploadManager.NewUpload(dfa.Hash, "file", dfa.NetworkName, dfa.UploaderAddress, dfa.HoldTimeInMonths)
			if err != nil {
				fmt.Println("error updating database ", err)
			}
			d.Ack(false)
			continue
		}
		// this isn't a new upload so we shall upload the database
		_, err = uploadManager.UpdateUpload(dfa.HoldTimeInMonths, dfa.UploaderAddress, dfa.Hash, dfa.NetworkName)
		if err != nil {
			fmt.Println("error updating database ", err)
			d.Ack(false)
			continue
		}
		d.Ack(false)
	}
}
