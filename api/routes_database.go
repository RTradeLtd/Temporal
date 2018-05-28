package api

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/gin-gonic/gin"
)

// RunTestGarbageCollection is used to run a test
// of our garbage collector
func RunTestGarbageCollection(c *gin.Context) {
	db := database.OpenDBConnection()
	um := models.NewUploadManager(db)
	deletedUploads := um.RunTestDatabaseGarbageCollection()
	c.JSON(http.StatusOK, gin.H{"deleted": deletedUploads})
}

// GetUploadsFromDatabase is used to read a list of uploads from our database
// TODO: cleanup
func GetUploadsFromDatabase(c *gin.Context) {
	// open a connection to the database
	db := database.OpenDBConnection()
	// create an upload manager interface
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
// TODO: cleanup
func GetUploadsForAddress(c *gin.Context) {
	// open connection to the database
	db := database.OpenDBConnection()
	// establish a new upload manager
	um := models.NewUploadManager(db)
	// fetch all uploads for that address
	uploads := um.GetUploadsForAddress(c.Param("address"))
	if uploads == nil {
		um.DB.Close()
		c.JSON(http.StatusNotFound, nil)
		return
	}
	um.DB.Close()
	c.JSON(http.StatusFound, gin.H{"uploads": uploads})
}
