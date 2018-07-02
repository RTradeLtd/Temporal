package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessDatabaseFileAdds is used to process database file add messages
func ProcessDatabaseFileAdds(msgs <-chan amqp.Delivery, db *gorm.DB) {
	cm := rtfs_cluster.Initialize()
	for d := range msgs {
		if d.Body != nil {
			if d.Body != nil {
				dfa := DatabaseFileAdd{}
				upload := models.Upload{}
				// unmarshal the message body into the dfa struct
				err := json.Unmarshal(d.Body, &dfa)
				if err != nil {
					d.Ack(false)
					continue
				}
				// convert the int64 to an int. We need to make sure to add a check that we won't overflow
				holdTime, err := strconv.Atoi(fmt.Sprintf("%v", dfa.HoldTimeInMonths))
				if err != nil {
					d.Ack(false)
					continue
				}
				// we will take the current time, and add the number of months to get the date
				// that we will garbage collect this from our repo
				gcd := time.Now().AddDate(0, holdTime, 0)
				upload.Hash = dfa.Hash
				upload.HoldTimeInMonths = dfa.HoldTimeInMonths
				upload.Type = "file"
				upload.UploadAddress = dfa.UploaderAddress
				upload.NetworkName = dfa.NetworkName
				upload.GarbageCollectDate = gcd
				lastUpload := models.Upload{
					Hash: dfa.Hash,
				}
				// retrieve the last upload matching this hash.
				// this upload will have the latest Garbage Collect Date
				db.Last(&lastUpload)
				// check the garbage collect dates, if the current upload to be pinned will be
				// GCd before the latest one from the database, we will skip it
				// however if it will be GCd at a later date, we will keep it
				// and update the database
				if lastUpload.GarbageCollectDate.Unix() >= upload.GarbageCollectDate.Unix() {
					d.Ack(false)
					// skip the rest of the message, preventing a database record from being created
					continue
				}
				upload.UploaderAddresses = append(lastUpload.UploaderAddresses, dfa.UploaderAddress)
				// we have a valid upload request, so lets store it to the database
				db.Create(&upload)
				go func() {
					decodedHash, err := cm.DecodeHashString(dfa.Hash)
					if err != nil {
						fmt.Println("error decoding hash ", err)
						return
					}
					err = cm.Pin(decodedHash)
					if err != nil {
						fmt.Println("error pinning to cluster")
					}
				}()
				d.Ack(false)
			}
		}
	}
}

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
			upload.NetworkName = dpa.NetworkName
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
