package queue

import (
	"github.com/RTradeLtd/database/models"
	log "github.com/sirupsen/logrus"
)

// refundCredits is used to refund a users credits. We do not check for errors,
// as these are logged and manually corrected if they occur
func (qm *Manager) refundCredits(username, callType string, cost float64) {
	if cost == 0 {
		return
	}
	um := models.NewUserManager(qm.db)
	if _, err := um.AddCredits(username, cost); err != nil {
		qm.logger.WithFields(log.Fields{
			"service":   qm.Service,
			"user":      username,
			"call_type": callType,
			"error":     err.Error(),
		}).Error("unable to refund user credits")
	}
}
