package api

import (
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/payment_server"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
)

/*
	These are routes related to tineracting with the blockhcain
*/

var method uint8

// RegisterPayment is used to register a payment with the payments contract
func RegisterPayment(c *gin.Context) {
	contextCopy := c.Copy()
	ethAddress := GetAuthenticatedUserFromContext(contextCopy)

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

	retentionPeriodInMonthsInt, err := strconv.ParseInt(retentionPeriodInMonths, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to convert retention period in months string to int",
		})
		return
	}

	paymentMethod, exists := contextCopy.GetPostForm("payment_method")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "payment_method post form param does not exist",
		})
		return
	}

	switch paymentMethod {
	case "rtc":
		method = 0
	case "eth":
		method = 1
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "not a valid payment method, must be eth or rtc",
		})
		return
	}

	rtfs := rtfs.Initialize("")

	costUsdFloat, err := utils.CalculatePinCost(contentHash, retentionPeriodInMonthsInt, rtfs.Shell)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unable to calcualte pin cost %s", err.Error()),
		})
		return
	}
	costUsdBigInt := utils.FloatToBigInt(costUsdFloat)
	chargeAmountInWei := utils.ConvertNumberToBaseWei(costUsdBigInt)

	dbPass := contextCopy.MustGet("db_pass").(string)
	dbURL := contextCopy.MustGet("db_url").(string)
	dbUser := contextCopy.MustGet("db_user").(string)
	db, err := database.OpenDBConnection(dbPass, dbURL, dbUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to open connection to database",
		})
		return
	}
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

	tx, err := pm.RegisterPaymentForUploader(ethAddress, contentHash, big.NewInt(retentionPeriodInMonthsInt), chargeAmountInWei, method, mqURL)
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
