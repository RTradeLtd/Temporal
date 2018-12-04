package v2

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/database/models"
	"github.com/gin-gonic/gin"
)

// GetUploadsFromDatabase is used to read a list of uploads from our database
func (api *API) getUploadsFromDatabase(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, eh.UnAuthorizedAdminAccess)
		return
	}
	um := models.NewUploadManager(api.dbm.DB)
	// fetch the uploads
	uploads, err := um.GetUploads()
	if err != nil {
		api.LogError(err, eh.UploadSearchError)(c, http.StatusInternalServerError)
		return
	}
	api.LogInfo("all uploads from database requested")
	Respond(c, http.StatusOK, gin.H{"response": uploads})
}

// GetUploadsForUser is used to read a list of uploads from a particular user name
// If not called by admin  admin, will retrieve all uploads for the current authenticated user
func (api *API) getUploadsForUser(c *gin.Context) {
	um := models.NewUploadManager(api.dbm.DB)
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}

	var queryUser string
	if err = api.validateAdminRequest(username); err == nil {
		queryUser = c.Param("user")
	} else {
		queryUser = username
	}
	// fetch all uploads for that address
	uploads, err := um.GetUploadsForUser(queryUser)
	if err != nil {
		api.LogError(err, eh.UploadSearchError)(c, http.StatusInternalServerError)
		return
	}
	api.LogInfo("specific uploads from database requested")
	Respond(c, http.StatusOK, gin.H{"response": uploads})
}
