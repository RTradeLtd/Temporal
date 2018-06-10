package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessDatabasePinAdds is used to process database pin adds
func ProcessDatabasePinAdds(msgs <-chan amqp.Delivery, db *gorm.DB) {
	for d := range msgs {
		if d.Body != nil {
			dpa := DatabasePinAdd{}
			upload := models.Upload{}
			log.Printf("receive a message: %s", d.Body)
			// unmarshal the message into the struct
			// if it can't be decoded into dpa struct, acknowledge message receival and continue to the nextm essage
			err := json.Unmarshal(d.Body, &dpa)
			// make this system more robust
			if err != nil {
				d.Ack(false)
				continue
			}
			upload.Hash = dpa.Hash
			upload.HoldTimeInMonths = dpa.HoldTimeInMonths
			upload.Type = "pin"
			upload.UploadAddress = dpa.UploaderAddress
			// get current time
			currTime := time.Now()
			// get the hold time from in64 and convert to int
			holdTime, err := strconv.Atoi(fmt.Sprint(dpa.HoldTimeInMonths))
			if err != nil {
				d.Ack(false)
				continue
			}
			// get the date the file wiill be garbage collected by adding the number of months
			gcd := currTime.AddDate(0, holdTime, 0)
			lastUpload := models.Upload{
				Hash: dpa.Hash,
			}
			db.Last(&lastUpload)
			// check to see whether or not the file will be garbage collected before the last upload
			// if so we'll skip
			if lastUpload.GarbageCollectDate.Unix() >= gcd.Unix() {
				fmt.Println("skipping since we already have an instance that will be GC'd later")
				d.Ack(false)
				continue
			}
			upload.UploaderAddresses = append(lastUpload.UploaderAddresses, dpa.UploaderAddress)
			upload.GarbageCollectDate = gcd
			db.Create(&upload)
			// submit message acknowledgement
			d.Ack(false)
		}
	}
}
