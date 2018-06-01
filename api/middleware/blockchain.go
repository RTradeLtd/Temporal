package middleware

import "github.com/gin-gonic/gin"

func BlockchainMiddleware(useIPC bool, ethKey, ethPass string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("use_ipc", useIPC)
		c.Set("eth_account", [2]string{ethKey, ethPass})
		// execute any pending handlers
		c.Next()
	}
}
