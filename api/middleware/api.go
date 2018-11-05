package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

/*
api middleware is used to secure access to the api
*/

var user models.User
var nilTime time.Time

// APIRestrictionMiddleware is used to restrict access to API calls
func APIRestrictionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		username := claims["id"]
		db.Where("user_name = ?", username).First(&user)
		if user.CreatedAt == nilTime {
			c.AbortWithError(http.StatusBadRequest, errors.New("invalid user account"))
			return
		}
		if !user.APIAccess {
			c.AbortWithError(http.StatusForbidden, errors.New("unauthorized api access"))
			return
		}
		c.Next()
	}
}
