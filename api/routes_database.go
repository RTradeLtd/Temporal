package api

import (
	"net/http"

	"github.com/RTradeLtd/database/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var dev = false

// GetUploadsFromDatabase is used to read a list of uploads from our database
func (api *API) getUploadsFromDatabase(c *gin.Context) {
	authenticatedUser := GetAuthenticatedUserFromContext(c)
	if authenticatedUser != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	um := models.NewUploadManager(api.dbm.DB)
	// fetch the uplaods
	uploads, err := um.GetUploads()
	if err != nil {
		api.LogError(err, UploadSearchError)(c, http.StatusInternalServerError)
		return
	}
	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    authenticatedUser,
	}).Info("all uploads from database requested")
	Respond(c, http.StatusOK, gin.H{"response": uploads})
}

// GetUploadsForAddress is used to read a list of uploads from a particular eth address
// If not called by admin  admin, will retrieve all uploads for the current authenticated user
func (api *API) getUploadsForAddress(c *gin.Context) {
	var queryUser string
	um := models.NewUploadManager(api.dbm.DB)
	user := GetAuthenticatedUserFromContext(c)
	if user == AdminAddress {
		queryUser = c.Param("user")
	} else {
		queryUser = user
	}
	// fetch all uploads for that address
	uploads, err := um.GetUploadsForUser(queryUser)
	if err != nil {
		api.LogError(err, UploadSearchError)(c, http.StatusInternalServerError)
		return
	}

	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    user,
	}).Info("specific uploads from database requested")

	Respond(c, http.StatusOK, gin.H{"response": uploads})
}
