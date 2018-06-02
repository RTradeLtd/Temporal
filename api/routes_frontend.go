package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
Contains routes used for frontend operation
*/

// SubmitPaymentRegistration is used to submit a payment registration
// request by an authenticated user
func SubmitPaymentRegistration(c *gin.Context) {
	contextCopy := c.Copy()
	uploaderAddress, exists := contextCopy.GetPostForm("eth_address")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "eth_address post form does not exist",
		})
		return
	}
	holdTime, exists := contextCopy.GetPostForm("hold_time")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "hold_time post form does not exist",
		})
		return
	}
	contentHash, exists := contextCopy.GetPostForm("content_hash")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "content_hash post form does not exist",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg": fmt.Sprint(uploaderAddress, holdTime, contentHash),
	})
}
