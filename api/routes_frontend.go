package api

import (
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/gin-gonic/gin"
)

/*
Contains routes used for frontend operation
*/

// SubmitPinPaymentRegistration is used to submit a payment registration
// request by an authenticated user
func SubmitPinPaymentRegistration(c *gin.Context) {
	var method uint8
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
	paymentMethod, exists := contextCopy.GetPostForm("payment_method")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "payment_method post form does not exist",
		})
		return
	}
	switch paymentMethod {
	case "rtc":
		method = 0
		break
	case "eth":
		method = 1
		break
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "provided payment does not exist, valid parameters are rtc or eth",
		})
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to convert hold time to int",
		})
		return
	}
	mqURL := c.MustGet("mq_url").(string)
	ppr := queue.PinPaymentRequest{
		UploaderAddress:  uploaderAddress,
		CID:              contentHash,
		HoldTimeInMonths: holdTimeInt,
		Method:           method,
	}

	qm, err := queue.Initialize(queue.PinPaymentRequestQueue, mqURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	err = qm.PublishMessage(ppr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "payment registration request sent",
	})
}
