package api

import (
	"math/big"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/server"
	"github.com/gin-gonic/gin"
)

/*
	These are routes related to tineracting with the blockhcain
*/

// RegisterRtcPayment is used to register an RTC payment with
// our smart contracts
func RegisterRtcPayment(c *gin.Context) {
	user := GetAuthenticatedUserFromContext(c)
	if user != AdminAddress {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "unauthorized access to register rtc payment",
		})
		return
	}
	ethAddress, exists := c.GetPostForm("eth_address")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "eth_address post form param not found",
		})
		return
	}
	contentHash, exists := c.GetPostForm("content_hash")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "content_hash post form param not found",
		})
		return
	}
	retentionPeriodInMonths, exists := c.GetPostForm("retention_period_in_months")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "retention_period_in_months post form param not found",
		})
	}
	chargeAmountInWei, exists := c.GetPostForm("charge_amount_in_wei")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "charge_amount_in_wei post form param not found",
		})
		return
	}
	chargeAmountInWeiInt, err := strconv.ParseInt(chargeAmountInWei, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to convert charge amount in wei to string",
		})
		return
	}
	retentionPeriodInMonthsInt, err := strconv.ParseInt(retentionPeriodInMonths, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to convert retention period in months string to int",
		})
		return
	}
	mqURL := c.MustGet("mq_conn_url").(string)
	sm := server.Initialize(true)
	tx, err := sm.RegisterPaymentForUploader(ethAddress, contentHash, big.NewInt(retentionPeriodInMonthsInt), big.NewInt(chargeAmountInWeiInt), 0, mqURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": "unable to process payment ofr uploader",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"tx_hash": tx.Hash().String(),
	})
	return
}
