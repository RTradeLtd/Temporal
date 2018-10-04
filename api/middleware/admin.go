package middleware

import (
	"errors"
	"net/http"

	"github.com/RTradeLtd/Temporal/models"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// AdminRestrictionMiddleware is used to lock down admin protected routes
func AdminRestrictionMiddleware(db *gorm.DB, adminUser string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		username := claims["id"].(string)
		um := models.NewUserManager(db)
		isAdmin, err := um.CheckIfAdmin(username)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, errors.New("error running middleware"))
			return
		}
		if !isAdmin {
			c.AbortWithError(http.StatusForbidden, errors.New("user is not an admin"))
			return
		}
		c.Next()
	}
}
