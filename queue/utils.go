package queue

import (
	"github.com/RTradeLtd/Temporal/models"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// refundCredits is used to refund a users credits. We do not check for errors,
// as these are logged and manually corrected if they occur
func (qm *QueueManager) refundCredits(username, callType string, cost float64, db *gorm.DB) {
	if cost == 0 {
		return
	}
	um := models.NewUserManager(db)
	if _, err := um.AddCredits(username, cost); err != nil {
		qm.Logger.WithFields(log.Fields{
			"service":   qm.Service,
			"user":      username,
			"call_type": callType,
			"error":     err.Error(),
		}).Error("unable to refund user credits")
	}
}
