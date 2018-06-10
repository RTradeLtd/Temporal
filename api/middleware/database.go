package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

/*
	Used for common connections to the database
*/

// DatabaseMiddleware is used for connections to the database
func DatabaseMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		// only call this inside middleware
		// it's purpose is to execute any pending handlers
		c.Next()
	}
}
