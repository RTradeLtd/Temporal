package api

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/gin-gonic/gin"
)

var dev = false

// GetUploadsFromDatabase is used to read a list of uploads from our database
func (api *API) getUploadsFromDatabase(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	um := models.NewUploadManager(api.dbm.DB)
	// fetch the uplaods
	uploads, err := um.GetUploads()
	if err != nil {
		api.LogError(err, UploadSearchError)(c, http.StatusInternalServerError)
		return
	}
	api.LogInfo("all uploads from database requested")
	Respond(c, http.StatusOK, gin.H{"response": uploads})
}

// GetUploadsForUser is used to read a list of uploads from a particular user name
// If not called by admin  admin, will retrieve all uploads for the current authenticated user
func (api *API) getUploadsForUser(c *gin.Context) {
	var queryUser string
	um := models.NewUploadManager(api.dbm.DB)
	username := GetAuthenticatedUserFromContext(c)
	err := api.validateAdminRequest(username)
	if err == nil {
		queryUser = c.Param("user")
	} else {
		queryUser = username
	}
	// fetch all uploads for that address
	uploads, err := um.GetUploadsForUser(queryUser)
	if err != nil {
		api.LogError(err, UploadSearchError)(c, http.StatusInternalServerError)
		return
	}
	api.LogInfo("specific uploads from database requested")
	Respond(c, http.StatusOK, gin.H{"response": uploads})
}
