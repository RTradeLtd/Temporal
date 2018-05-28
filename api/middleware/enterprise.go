package middleware

import (
	"errors"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

/*

This is a WIP
The enterprise middleware will only process pin requests, and will skip
all other processing (ie, uploader, etc...) and requires its own database table
*/

// EnterpriseMiddleware is used to circumvent the non-enterprise processing
// including checking for user uploader addresses, hold times, etc...
// since the point of the enterprise API is access to the underlying temporal components
// this middleware handles updating the database
func EnterpriseMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		hash := c.Param("hash")
		if hash == "" {
			c.Error(errors.New("invalid hash"))
			return
		}
		upload := models.Upload{Hash: hash}
		db.Create(&upload)
		c.Next()
	}
}
