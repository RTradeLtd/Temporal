package middleware

import (
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

// NewSecWare is used to initialize our security middleware
func NewSecWare(devMode bool) gin.HandlerFunc {
	config := secure.DefaultConfig()
	config.IsDevelopment = devMode
	return secure.New(config)
}
