package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware is used to load our CORS handling logic
// TODO: need to cleanup and restrict origins, the all origin allow
// and credentials disable is a temporary measure
func CORSMiddleware() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowCredentials = false
	corsConfig.AddAllowHeaders("cache-control", "Access-Control-Allow-Headers", "Authorization", "Content-Type", "Access-Control-Allow-Origin", "Access-Control-Request-Headers")
	return cors.New(corsConfig)
}
