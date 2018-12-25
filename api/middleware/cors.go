package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware is used to load our CORS handling logic
func CORSMiddleware() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	// update the allowed origins (covers development website, and the production website)
	corsConfig.AllowOrigins = []string{"https://nuts.rtradetechnologies.com:6771", "https://web.temporal.cloud:6771"}
	// allow the DELETE method, allowed methods are now
	// DELETE GET POST PUT HEAD
	corsConfig.AddAllowMethods("DELETE")
	corsConfig.AddAllowHeaders("cache-control", "Authorization", "Content-Type")
	return cors.New(corsConfig)
}
