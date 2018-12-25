package middleware

import (
	"errors"
	"net/http"

	"github.com/RTradeLtd/database/models"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// APIRestrictionMiddleware is used to restrict access to API calls
func APIRestrictionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		username := claims["id"]
		var user models.User
		if check := db.Where("user_name = ?", username).First(&user); check.Error != nil {
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
