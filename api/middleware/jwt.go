package middleware

import (
	"time"

	"github.com/RTradeLtd/database/v2/models"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
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
		// userId will be either the username or email address
		Authenticator: func(userId string, password string, c *gin.Context) (string, bool) {
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
			// fixes https://github.com/RTradeLtd/Temporal/issues/405
			// regardless of whether or not they are providing username or email
			// always return the username
			usr, err := userManager.FindByUserName(userId)
			if err != nil {
				usr, err = userManager.FindByEmail(userId)
				if err != nil {
					lAuth.Warn("failed to find user", "error", err)
					return "", false
				}
			}
			// email enabled implies they have verified their email
			if !usr.EmailEnabled {
				return "", false
			}
			lAuth.Info("successful login", "username", usr.UserName)
			return usr.UserName, true
		},
		Authorizator: func(userId string, c *gin.Context) bool {
			// as a final security step, ensure that we can find the user in our database
			userManager := models.NewUserManager(db)
			usr, err := userManager.FindByUserName(userId)
			if err != nil {
				return false
			}
			if usr.EmailEnabled && usr.AccountEnabled {
				return true
			}
			return false
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
