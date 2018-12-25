package middleware

import (
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

// NewSecWare is used to initialize our security middleware
func NewSecWare() gin.HandlerFunc {
	config := secure.DefaultConfig()
	config.AllowedHosts = []string{"https://nuts.rtradetechnologies.com:6771", "https://web.temporal.cloud:6771"}
	config.SSLRedirect = true
	return secure.New(config)
}
