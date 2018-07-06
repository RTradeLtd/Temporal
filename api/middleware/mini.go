package middleware

import (
	"github.com/gin-gonic/gin"
)

/*
Minio Middleware is used to handle our connection to minio
*/

func MINIMiddleware(accessKey, secretKey, endpoint string, secure bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("minio_credentials", map[string]string{"access_key": accessKey, "secret_key": secretKey})
		c.Set("minio_secure", secure)
		c.Set("minio_endpoint", endpoint)
		c.Next()
	}
}
