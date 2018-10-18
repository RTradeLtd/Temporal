package queue

import log "github.com/sirupsen/logrus"

// LogError is a wrapper used to log an error message
func (qm *QueueManager) LogError(err error, message string) {
	qm.Logger.WithFields(log.Fields{
		"service": qm.Service,
		"error":   err.Error(),
	}).Error(message)
}

// LogInfo is a wrapper used to log an info message
func (qm *QueueManager) LogInfo(message string) {
	qm.Logger.WithFields(log.Fields{
		"service": qm.Service,
	}).Info(message)
}
