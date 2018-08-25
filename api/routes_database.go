package api

import (
	"errors"
	"net/http"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/gin-gonic/gin"
)

var dev = false

// GetUploadsFromDatabase is used to read a list of uploads from our database
// only usable by admin
func (api *API) getUploadsFromDatabase(c *gin.Context) {
	authenticatedUser := GetAuthenticatedUserFromContext(c)
	if authenticatedUser != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	um := models.NewUploadManager(api.DBM.DB)
	// fetch the uplaods
	uploads := um.GetUploads()
	if uploads == nil {
		FailOnError(c, errors.New("no uploads found"))
		return
	}
	c.JSON(http.StatusFound, gin.H{"uploads": uploads})
}

// GetUploadsForAddress is used to read a list of uploads from a particular eth address
// If not admin, will retrieve all uploads for the current context account
func (api *API) getUploadsForAddress(c *gin.Context) {
	var queryUser string
	um := models.NewUploadManager(api.DBM.DB)
	user := GetAuthenticatedUserFromContext(c)
	if user == AdminAddress {
		queryUser = c.Param("user")
	} else {
		queryUser = user
	}
	// fetch all uploads for that address
	uploads := um.GetUploadsForUser(queryUser)
	if uploads == nil {
		FailOnError(c, errors.New("no uploads found"))
		return
	}
	c.JSON(http.StatusFound, gin.H{"uploads": uploads})
}
