package tns

import (
	log "github.com/sirupsen/logrus"
)

// LogError is a helper used to log an error
func (m *Manager) LogError(err error, message string, fields ...interface{}) {
	// create the base entry
	entry := m.l.WithFields(log.Fields{
		"service": m.service,
	})
	// add additional fields
	if fields != nil {
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				entry = entry.WithField(fields[i].(string), fields[i+1])
			}
		}
	}
	if err == nil {
		entry.Error(message)
	} else {
		entry.WithField("error", err.Error()).Error(message)
	}
}

// LogInfo is a helper used to log an informational message
func (m *Manager) LogInfo(message ...interface{}) {
	m.l.WithFields(log.Fields{
		"service": m.service,
	}).Info(message...)
}
