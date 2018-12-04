package v2

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// LogError is a wrapper used by the API to handle logging of errors. Returns a
// callback to also fail a gin context with an optional status code, which
// defaults to http.StatusInternalServerError. Fields is an optional set of
// params provided in pairs, where the first of a pair is the key, and the second
// is the value
func (api *API) LogError(err error, message string, fields ...interface{}) func(c *gin.Context, code ...int) {
	// create base entry
	entry := api.l.WithFields(log.Fields{
		"service": api.service,
	})

	// add additional fields
	if fields != nil {
		for i := 0; i < len(fields); i += 2 {
			if i+1 < len(fields) {
				entry = entry.WithField(fields[i].(string), fields[i+1])
			}
		}
	}

	// write log
	if err == nil {
		entry.Error(message)
	} else {
		entry.WithField("error", err.Error()).Error(message)
	}

	// return utility callback
	if message == "" && err != nil {
		return func(c *gin.Context, code ...int) { Fail(c, err, code...) }
	}
	return func(c *gin.Context, code ...int) { FailWithMessage(c, message, code...) }
}

// LogInfo is a wrapper used by the API to handle simple info logs
func (api *API) LogInfo(message ...interface{}) {
	api.l.WithFields(log.Fields{
		"service": api.service,
	}).Info(message...)
}

// LogDebug is a wrapper used by the API to handle simple debug logs
func (api *API) LogDebug(message ...interface{}) {
	api.l.WithFields(log.Fields{
		"service": api.service,
	}).Debug(message...)
}

// LogWithUser creates entry context with user
func (api *API) LogWithUser(user string) *log.Entry {
	return api.l.WithFields(log.Fields{
		"service": api.service,
		"user":    user,
	})
}
