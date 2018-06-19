package middleware

import "github.com/gin-gonic/gin"

func AWSMiddleware(key, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("aws_key", key)
		c.Set("aws_secret", secret)
		// execute any pending handlers
		c.Next()
	}
}
