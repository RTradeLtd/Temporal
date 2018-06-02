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

func APIRestrictionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		ethAddress := claims["id"]
		db.Where("eth_address = ?", ethAddress).First(&user)
		if user.CreatedAt == nilTime {
			c.AbortWithError(http.StatusBadRequest, errors.New("invalid user account"))
		}
		if !user.APIAccess {
			c.AbortWithError(http.StatusForbidden, errors.New("unauthorized api access"))
		}
		c.Next()
	}
}
