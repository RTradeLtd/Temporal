package api

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/payment_server"
	"github.com/gin-gonic/gin"
)

/*
	These are routes related to tineracting with the blockhcain
*/

// RegisterRtcPayment is used to register an RTC payment with
// our smart contracts
func RegisterRtcPayment(c *gin.Context) {
	contextCopy := c.Copy()
	ethAddress, exists := contextCopy.GetPostForm("eth_address")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "eth_address post form param not found",
		})
		return
	}
	contentHash, exists := contextCopy.GetPostForm("content_hash")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "content_hash post form param not found",
		})
		return
	}
	retentionPeriodInMonths, exists := contextCopy.GetPostForm("retention_period_in_months")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "retention_period_in_months post form param not found",
		})
		return
	}
	chargeAmountInWei, exists := contextCopy.GetPostForm("charge_amount_in_wei")
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
	dbPass := contextCopy.MustGet("db_pass").(string)
	dbURL := contextCopy.MustGet("db_url").(string)
	dbUser := contextCopy.MustGet("db_user").(string)
	db := database.OpenDBConnection(dbPass, dbURL, dbUser)
	mqURL := contextCopy.MustGet("mq_conn_url").(string)
	ethAccount := contextCopy.MustGet("eth_account").([2]string) // 0 = key, 1 = pass
	// since we aren't interacting with any contract events we dont need IPC
	pm, err := payment_server.NewPaymentManager(false, ethAccount[0], ethAccount[1], db)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unable to create payment manager %s", err.Error()),
		})
		return
	}

	tx, err := pm.RegisterPaymentForUploader(ethAddress, contentHash, big.NewInt(retentionPeriodInMonthsInt), big.NewInt(chargeAmountInWeiInt), 0, mqURL)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": fmt.Sprintf("unable to process payment for uploader %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"tx_hash": tx.Hash().String(),
	})
	return
}

func RegisterEthPayment(c *gin.Context) {
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
	dbPass := c.MustGet("db_pass").(string)
	dbURL := c.MustGet("db_url").(string)
	dbUser := c.MustGet("db_user").(string)
	db := database.OpenDBConnection(dbPass, dbURL, dbUser)
	mqURL := c.MustGet("mq_conn_url").(string)
	ethAccount := c.MustGet("eth_account").([2]string) // 0 = key, 1 = pass
	pm, err := payment_server.NewPaymentManager(false, ethAccount[0], ethAccount[1], db)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unable to create payment manager %s", err.Error()),
		})
		return
	}

	tx, err := pm.RegisterPaymentForUploader(ethAddress, contentHash, big.NewInt(retentionPeriodInMonthsInt), big.NewInt(chargeAmountInWeiInt), 1, mqURL)
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
