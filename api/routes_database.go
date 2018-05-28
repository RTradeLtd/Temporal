package api

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// RunTestGarbageCollection is used to run a test
// of our garbage collector
func RunTestGarbageCollection(c *gin.Context) {
	user := GetAuthenticatedUserFromContext(c)
	if user != AdminAddress {
		c.JSON(http.StatusForbidden, gin.H{"error": "unauthorized access to test garbage collect"})
	}
	db := c.MustGet("db_connection").(*gorm.DB)
	um := models.NewUploadManager(db)
	deletedUploads := um.RunTestDatabaseGarbageCollection()
	c.JSON(http.StatusOK, gin.H{"deleted": deletedUploads})
}

// GetUploadsFromDatabase is used to read a list of uploads from our database
// only usable by admin
func GetUploadsFromDatabase(c *gin.Context) {
	user := GetAuthenticatedUserFromContext(c)
	if user != AdminAddress {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden access, not an admin this will be reported"})
		// trigger error report
		return
	}
	// open a connection to the database
	db := c.MustGet("db_connection").(*gorm.DB)
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
	// open a connection to the database
	db := c.MustGet("db_connection").(*gorm.DB)
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
