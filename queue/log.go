package queue

import log "github.com/sirupsen/logrus"

// LogError is a wrapper used to log an error message
func (qm *Manager) LogError(err error, message string, fields ...interface{}) {
	// base entry
	entry := qm.Logger.WithFields(log.Fields{
		"service": qm.Service,
	})

	// parse additional fields if there are any
	if fields != nil {
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				entry = entry.WithField(fields[i].(string), fields[i+1])
			}
		}
	}

	// write the actual log
	if err == nil {
		entry.Error(message)
	} else {
		entry.WithField("error", err.Error()).Error(message)
	}
}

// LogInfo is a wrapper used to log an info message
func (qm *Manager) LogInfo(message string) {
	qm.Logger.WithFields(log.Fields{
		"service": qm.Service,
	}).Info(message)
}
