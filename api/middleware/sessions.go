package middleware

import (
	"github.com/RTradeLtd/Temporal/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// https://skarlso.github.io/2016/11/02/google-signin-with-go-part2/
// https://github.com/gin-contrib/sessions/issues/42

// SessionMiddleware is used to handle our sessions
func SessionMiddleware(cfg *config.TemporalConfig) gin.HandlerFunc {
	cookieStore := cookie.NewStore([]byte(cfg.API.Sessions.AuthKey), []byte(cfg.API.Sessions.EncryptionKey))
	cookieStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 1 day
		Secure:   true,
		HttpOnly: false,
	})
	return sessions.Sessions("mysession", cookieStore)
}
