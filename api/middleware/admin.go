package middleware

import (
	"errors"
	"net/http"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// AdminRestrictionMiddleware is used to lock down admin protected routes
func AdminRestrictionMiddleware(db *gorm.DB, adminAdress string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		ethAddress := claims["id"]
		if ethAddress != adminAdress {
			c.AbortWithError(http.StatusForbidden, errors.New("user is not an admin"))
			return
		}
		c.Next()
	}
}
