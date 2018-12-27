package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	allowedOrigins = []string{"https://temporal.cloud"}
)

// CORSMiddleware is used to load our CORS handling logic
func CORSMiddleware() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	// configure allowed origins
	corsConfig.AllowOrigins = allowedOrigins
	// allow the DELETE method, allowed methods are now
	// DELETE GET POST PUT HEAD
	corsConfig.AddAllowMethods("DELETE")
	corsConfig.AddAllowHeaders("cache-control", "Authorization", "Content-Type")
	return cors.New(corsConfig)
}
