package queue

import (
	"github.com/RTradeLtd/database/v2/models"
)

// refundCredits is used to refund a users credits. We do not check for errors,
// as these are logged and manually corrected if they occur
func (qm *Manager) refundCredits(username, callType string, cost float64) error {
	if cost == 0 {
		return nil
	}
	um := models.NewUserManager(qm.db)
	if _, err := um.AddCredits(username, cost); err != nil {
		qm.l.Errorw(
			"failed to refund user credits",
			"error", err.Error(),
			"user", username,
			"call_type", callType,
			"cost", cost)
		return err
	}
	return nil
}
