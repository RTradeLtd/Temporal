package api

import (
	"errors"
	"net/http"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var dev = false

// RunTestGarbageCollection is used to run a test
// of our garbage collector
func RunTestGarbageCollection(c *gin.Context) {
	if !dev {
		FailOnError(c, errors.New("attempting to run test database garbage colelction in non dev mode"))
		return
	}
	authenticatedUser := GetAuthenticatedUserFromContext(c)
	if authenticatedUser != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}
	um := models.NewUploadManager(db)
	deletedUploads, err := um.RunTestDatabaseGarbageCollection()
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": deletedUploads})
}

// RunDatabaseGarbageCollection is used to parse through our database
// and remove them from the database. Note we do not remove from the cluster here
// since that will be a pretty long call. We will have a script that parses through
// the database once every 24 hours, looking for any deleted pins, and removing them
// from the cluster
func RunDatabaseGarbageCollection(c *gin.Context) {
	authenticatedUser := GetAuthenticatedUserFromContext(c)
	if authenticatedUser != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}
	um := models.NewUploadManager(db)
	deletedUploads, err := um.RunDatabaseGarbageCollection()
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted_uploads": deletedUploads})
}

// GetUploadsFromDatabase is used to read a list of uploads from our database
// only usable by admin
func GetUploadsFromDatabase(c *gin.Context) {
	authenticatedUser := GetAuthenticatedUserFromContext(c)
	if authenticatedUser != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}
	um := models.NewUploadManager(db)
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
func GetUploadsForAddress(c *gin.Context) {
	var queryAddress string
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}

	um := models.NewUploadManager(db)
	user := GetAuthenticatedUserFromContext(c)
	if user == AdminAddress {
		queryAddress = c.Param("address")
	} else {
		queryAddress = user
	}
	// fetch all uploads for that address
	uploads := um.GetUploadsForAddress(queryAddress)
	if uploads == nil {
		FailOnError(c, errors.New("no uploads found"))
		return
	}
	c.JSON(http.StatusFound, gin.H{"uploads": uploads})
}
