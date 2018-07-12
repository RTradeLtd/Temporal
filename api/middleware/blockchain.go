package middleware

import "github.com/gin-gonic/gin"

func BlockchainMiddleware(useIPC bool, ethKey, ethPass string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("use_ipc", useIPC)
		m := make(map[string]string)
		m["keyFile"] = ethKey
		m["keyPass"] = ethPass
		c.Set("eth_account", m)
		// execute any pending handlers
		c.Next()
	}
}
