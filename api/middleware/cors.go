package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	// DefaultAllowedOrigins are the default allowed origins for the api, allowing for access
	// via both of our internet uplinks when using the web interface.
	DefaultAllowedOrigins = []string{"https://temporal.cloud", "https://backup.temporal.cloud"}
)

// CORSMiddleware is used to load our CORS handling logic
func CORSMiddleware(devMode bool, allowedOrigins []string) gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	if devMode {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = false
	} else {
		// configure allowed origins
		corsConfig.AllowOrigins = allowedOrigins
	}
	// allow the DELETE method, allowed methods are now
	// DELETE GET POST PUT HEAD
	corsConfig.AddAllowMethods("DELETE")
	corsConfig.AddAllowHeaders("cache-control", "Authorization", "Content-Type", "X-Request-ID")
	return cors.New(corsConfig)
}
