package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// https://skarlso.github.io/2016/11/02/google-signin-with-go-part2/
// https://github.com/gin-contrib/sessions/issues/42

func SessionMiddleware() gin.HandlerFunc {
	cookieStore := cookie.NewStore([]byte("secret_auth"), []byte("secret_encryption"))
	cookieStore.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 1 day
		Secure:   true,
		HttpOnly: false,
	})
	return sessions.Sessions("mysession", cookieStore)
}
