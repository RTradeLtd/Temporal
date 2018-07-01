package api

import (
	"errors"
	"math/big"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/payment_server"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

/*
	These are routes related to tineracting with the blockhcain
*/

var method uint8

// RegisterPayment is used to register a payment with the payments contract
func RegisterPayment(c *gin.Context) {
	contextCopy := c.Copy()
	ethAddress := GetAuthenticatedUserFromContext(contextCopy)
	// only allow the admin to access this function
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access to admin route")
		return
	}
	contentHash, exists := contextCopy.GetPostForm("content_hash")
	if !exists {
		FailNoExistPostForm(c, "content_hash")
		return
	}
	retentionPeriodInMonths, exists := contextCopy.GetPostForm("retention_period_in_months")
	if !exists {
		FailNoExistPostForm(c, "retention_period_in_months")
		return
	}

	retentionPeriodInMonthsInt, err := strconv.ParseInt(retentionPeriodInMonths, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}

	paymentMethod, exists := contextCopy.GetPostForm("payment_method")
	if !exists {
		FailNoExistPostForm(c, "payment_method")
		return
	}

	switch paymentMethod {
	case "rtc":
		method = 0
	case "eth":
		method = 1
	default:
		FailOnError(c, errors.New("not a valid payment method, must be eth or rtc"))
		return
	}

	rtfs, err := rtfs.Initialize("", "")
	if err != nil {
		FailOnError(c, err)
		return
	}
	// calculate the pin cost, we do this by fetching the total object size and the number of months they want to store for
	costUsdFloat, err := utils.CalculatePinCost(contentHash, retentionPeriodInMonthsInt, rtfs.Shell)
	if err != nil {
		FailOnError(c, err)
		return
	}
	// format the data
	costUsdBigInt := utils.FloatToBigInt(costUsdFloat)
	chargeAmountInWei := utils.ConvertNumberToBaseWei(costUsdBigInt)

	// load relevant connections
	db, ok := contextCopy.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}
	mqURL := contextCopy.MustGet("mq_conn_url").(string)
	ethAccount := contextCopy.MustGet("eth_account").([2]string) // 0 = key, 1 = pass
	// since we aren't interacting with any contract events we dont need IPC
	pm, err := payment_server.NewPaymentManager(false, ethAccount[0], ethAccount[1], db)
	if err != nil {
		FailOnError(c, err)
		return
	}

	tx, err := pm.RegisterPaymentForUploader(ethAddress, contentHash, big.NewInt(retentionPeriodInMonthsInt), chargeAmountInWei, method, mqURL)
	if err != nil {
		FailOnError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"tx_hash": tx.Hash().String(),
	})
	return
}
