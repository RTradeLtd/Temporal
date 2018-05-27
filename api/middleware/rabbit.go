package middleware

/*
	When running go routines inside middleware, we must copy the context

*/
import "github.com/gin-gonic/gin"

// RabbitMQMiddleware is used to load common paremeters
// needed for connection, and using rabbitmq from within
// any api calls
func RabbitMQMiddleware(connectionURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("mq_conn_url", connectionURL)
		// only call this inside middleware
		// it's purpose is to execute any pending handlers
		c.Next()
	}
}
