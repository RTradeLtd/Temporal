package middleware

import (
	"time"

	"github.com/RTradeLtd/database/v2/models"
	"github.com/RTradeLtd/gorm"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Login is used to unmarshal a login in request so that we can parse it
type Login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// JwtConfigGenerate is used to generate our JWT configuration
func JwtConfigGenerate(jwtKey, realmName string, db *gorm.DB, l *zap.SugaredLogger) *jwt.GinJWTMiddleware {
	l = l.Named("jwt-middleware")
	authMiddleware := &jwt.GinJWTMiddleware{
		Realm:      realmName,
		Key:        []byte(jwtKey),
		Timeout:    time.Hour * 24,
		MaxRefresh: time.Hour * 24,
		Authenticator: func(userId string, password string, c *gin.Context) (string, bool) { // userId = username
			lAuth := l.With("user", userId)
			userManager := models.NewUserManager(db)
			validLogin, err := userManager.SignIn(userId, password)
			if err != nil {
				lAuth.Warn("bad sign in", "error", err)
				return userId, false
			}
			if !validLogin {
				lAuth.Warn("bad login")
				return userId, false
			}
			lAuth.Info("successful login")
			return userId, true
		},
		Authorizator: func(userId string, c *gin.Context) bool {
			// as a final security step, ensure that we can find the user in our database
			userManager := models.NewUserManager(db)
			if _, err := userManager.FindByUserName(userId); err != nil {
				return false
			}
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			l.Error("invalid login detected")
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},

		TokenLookup:   "header:Authorization",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	}

	return authMiddleware
}
