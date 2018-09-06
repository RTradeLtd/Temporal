package api

import (
	"fmt"
	"net/http"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var dev = false

// GetUploadsFromDatabase is used to read a list of uploads from our database
func (api *API) getUploadsFromDatabase(c *gin.Context) {
	authenticatedUser := GetAuthenticatedUserFromContext(c)
	if authenticatedUser != AdminAddress {
		msg := fmt.Sprintf("user %s attempted unauthorized access to get uploads from database admin route", authenticatedUser)
		api.Logger.Warn(msg)
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	um := models.NewUploadManager(api.DBM.DB)
	// fetch the uplaods
	uploads, err := um.GetUploads()
	if err != nil {
		msg := fmt.Sprintf("get uploads from database failed due to the following error: %s", err.Error())
		api.Logger.Error(msg)
		FailOnError(c, err)
		return
	}
	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    authenticatedUser,
	}).Info("all uploads from database requested")
	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": uploads,
	})
}

// GetUploadsForAddress is used to read a list of uploads from a particular eth address
// If not called by admin  admin, will retrieve all uploads for the current authenticated user
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
	uploads, err := um.GetUploadsForUser(queryUser)
	if err != nil {
		msg := fmt.Sprintf("get uploads from database for user %s failed due to the following error: %s", queryUser, err.Error())
		api.Logger.Error(msg)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    user,
	}).Info("specific uploads from database requested")

	c.JSON(http.StatusOK, gin.H{
		"code":     http.StatusOK,
		"response": uploads,
	})
}
