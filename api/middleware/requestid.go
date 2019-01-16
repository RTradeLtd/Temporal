package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID is used to insert a randomly generated
// uuid as a value to a X-Request-ID header
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Request-ID", uuid.New().String())
		c.Next()
	}
}
