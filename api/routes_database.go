package api

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/gin-gonic/gin"
)

var dev = false

// RunTestGarbageCollection is used to run a test
// of our garbage collector
func RunTestGarbageCollection(c *gin.Context) {
	if !dev {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "attempting to run test database garbage colelction in non dev mode",
		})
	}
	dbPass := c.MustGet("db_pass").(string)
	dbURL := c.MustGet("db_url").(string)
	db := database.OpenDBConnection(dbPass, dbURL)
	um := models.NewUploadManager(db)
	deletedUploads := um.RunTestDatabaseGarbageCollection()
	c.JSON(http.StatusOK, gin.H{"deleted": deletedUploads})
}

// RunDatabaseGarbageCollection is used to parse through our database
// and remove them from the database. Note we do not remove from the cluster here
// since that will be a pretty long call. We will have a script that parses through
// the database once every 24 hours, looking for any deleted pins, and removing them
// from the cluster
func RunDatabaseGarbageCollection(c *gin.Context) {
	dbPass := c.MustGet("db_pass").(string)
	dbURL := c.MustGet("db_url").(string)
	db := database.OpenDBConnection(dbPass, dbURL)
	um := models.NewUploadManager(db)
	deletedUploads := um.RunDatabaseGarbageCollection()
	c.JSON(http.StatusOK, gin.H{
		"deleted_uploads": deletedUploads,
	})
}

// GetUploadsFromDatabase is used to read a list of uploads from our database
// only usable by admin
func GetUploadsFromDatabase(c *gin.Context) {
	dbPass := c.MustGet("db_pass").(string)
	dbURL := c.MustGet("db_url").(string)
	db := database.OpenDBConnection(dbPass, dbURL)
	um := models.NewUploadManager(db)
	// fetch the uplaods
	uploads := um.GetUploads()
	if uploads == nil {
		um.DB.Close()
		c.JSON(http.StatusNotFound, nil)
		return
	}
	um.DB.Close()
	c.JSON(http.StatusFound, gin.H{"uploads": uploads})
}

// GetUploadsForAddress is used to read a list of uploads from a particular eth address
// If not admin, will retrieve all uploads for the current context account
func GetUploadsForAddress(c *gin.Context) {
	var queryAddress string
	dbPass := c.MustGet("db_pass").(string)
	dbURL := c.MustGet("db_url").(string)
	db := database.OpenDBConnection(dbPass, dbURL)
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
		um.DB.Close()
		c.JSON(http.StatusNotFound, nil)
		return
	}
	um.DB.Close()
	c.JSON(http.StatusFound, gin.H{"uploads": uploads})
}
