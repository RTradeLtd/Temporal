package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func FailNoExist(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": message,
	})
}
