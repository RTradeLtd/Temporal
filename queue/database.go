package queue

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

var nilTime time.Time

// ProcessDatabaseFileAdds is used to process database file add messages
func ProcessDatabaseFileAdds(msgs <-chan amqp.Delivery, db *gorm.DB) {
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
				lastUpload := models.Upload{}
				if check := db.Where("hash = ? AND network_name = ?", upload.Hash, upload.NetworkName).Last(&lastUpload); check.Error != nil && check.Error != gorm.ErrRecordNotFound {
					//TODO: add error handling
					fmt.Println("Error ", check.Error)
					d.Ack(false)
					continue
				}
				// check the garbage collect dates, if the current upload to be pinned will be
				// GCd before the latest one from the database, we will skip it
				// however if it will be GCd at a later date, we will keep it
				// and update the database
				if lastUpload.GarbageCollectDate.Unix() > upload.GarbageCollectDate.Unix() {
					upload.GarbageCollectDate = lastUpload.GarbageCollectDate
				}
				upload.UploaderAddresses = append(lastUpload.UploaderAddresses, dfa.UploaderAddress)
				fmt.Println("Saving in database")
				if check := db.Save(&upload); check.Error != nil {
					//TOOD add error handling
					fmt.Println("error ", check.Error)
					d.Ack(false)
					continue
				}
				fmt.Println("record saved")
			}
		}
	}
}

// ProcessDatabasePinAdds is used to process database file add messages
func ProcessDatabasePinAdds(msgs <-chan amqp.Delivery, db *gorm.DB) {
	for d := range msgs {
		if d.Body != nil {
			if d.Body != nil {
				dpa := DatabasePinAdd{}
				upload := models.Upload{}
				// unmarshal the message body into the dfa struct
				err := json.Unmarshal(d.Body, &dpa)
				if err != nil {
					d.Ack(false)
					continue
				}
				// convert the int64 to an int. We need to make sure to add a check that we won't overflow
				holdTime, err := strconv.Atoi(fmt.Sprintf("%v", dpa.HoldTimeInMonths))
				if err != nil {
					d.Ack(false)
					continue
				}
				// we will take the current time, and add the number of months to get the date
				// that we will garbage collect this from our repo
				gcd := time.Now().AddDate(0, holdTime, 0)
				upload.Hash = dpa.Hash
				upload.HoldTimeInMonths = dpa.HoldTimeInMonths
				upload.Type = "pin"
				upload.UploadAddress = dpa.UploaderAddress
				upload.NetworkName = dpa.NetworkName
				upload.GarbageCollectDate = gcd
				lastUpload := models.Upload{}
				if check := db.Where("hash = ? AND network_name = ?", upload.Hash, upload.NetworkName).Last(&lastUpload); check.Error != nil && check.Error != gorm.ErrRecordNotFound {
					//TODO: add error handling
					fmt.Println("Error ", check.Error)
					d.Ack(false)
					continue
				}
				// check the garbage collect dates, if the current upload to be pinned will be
				// GCd before the latest one from the database, we will skip it
				// however if it will be GCd at a later date, we will keep it
				// and update the database
				if lastUpload.GarbageCollectDate.Unix() > upload.GarbageCollectDate.Unix() {
					upload.GarbageCollectDate = lastUpload.GarbageCollectDate
				}
				upload.UploaderAddresses = append(lastUpload.UploaderAddresses, dpa.UploaderAddress)
				fmt.Println("Saving in database")
				if check := db.Save(&upload); check.Error != nil {
					//TOOD add error handling
					fmt.Println("error ", check.Error)
					d.Ack(false)
					continue
				}
				fmt.Println("record saved")
			}
		}
	}
}
