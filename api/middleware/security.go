package middleware

import (
	"github.com/gin-contrib/secure"
	"github.com/gin-gonic/gin"
)

// NewSecWare is used to initialize our security middleware
func NewSecWare(devMode bool) gin.HandlerFunc {
	config := secure.DefaultConfig()
	config.IsDevelopment = devMode
	config.ContentSecurityPolicy = "default-src 'self' https://checkout.stripe.com; connect-src https://checkout.stripe.com; frame-src https://checkout.stripe.com; script-src https://checkout.stripe.com; img-src https://*.stripe.com; object-src 'none'"
	return secure.New(config)
}
