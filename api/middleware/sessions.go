package middleware

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// https://skarlso.github.io/2016/11/02/google-signin-with-go-part2/

func SessionMiddleware(cfg *config.TemporalConfig) gin.HandlerFunc {
	cookieStore := cookie.NewStore([]byte("secret_auth"), []byte("secret_encryption"))
	cookieStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 1 day
		Secure:   true,
		HttpOnly: false,
	})
	return sessions.Sessions("mysession", cookieStore)
}

func t(c *gin.Context) {
	session := sessions.Default(c)
	session.
}

func GenerateToken32() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func GenerateToken64() string {
	b := make([]byte, 64)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
