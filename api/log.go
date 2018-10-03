package api

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// LogError is a wrapper used by the API to handle logging of errors. Returns a
// callback to also fail a gin context with an optional status code, which
// defaults to http.StatusInternalServerError
func (api *API) LogError(err error, message string) func(c *gin.Context, code ...int) {
	if err == nil {
		api.l.WithFields(log.Fields{
			"service": api.service,
		}).Error(message)
	} else {
		api.l.WithFields(log.Fields{
			"service": api.service,
			"error":   err.Error(),
		}).Error(message)
	}
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
