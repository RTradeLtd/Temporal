package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/utils"

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
	mqURL := c.MustGet("mq_conn_url").(string)
	manager := rtfs.Initialize("")
	pinCostUsd, err := utils.CalculatePinCost(contentHash, holdTimeInt, manager.Shell)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to calculate cost of pin",
		})
		return
	}
	var cost float64
	// TODO: use a money/currency library for the math.
	// this is a place holder
	switch method {
	case 0:
		rtcUSD := float64(0.125)
		cost = pinCostUsd / rtcUSD
	case 1:
		ethUSD, err := utils.RetrieveEthUsdPrice()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("error ", err),
			})
		}
		cost = pinCostUsd / ethUSD
		break
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid payment method, must be 0 or 1",
		})
		return
	}
	fmt.Println(cost)
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
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   "payment registration request sent",
		"cost_usd": pinCostUsd,
	})
}

// CalculatePinCost is used to calculate the cost of pinning something to temporal
func CalculatePinCost(c *gin.Context) {
	hash := c.Param("hash")
	holdTime := c.Param("holdtime")
	manager := rtfs.Initialize("")
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to convert hold time from string to int",
		})
		return
	}
	totalCost, err := utils.CalculatePinCost(hash, holdTimeInt, manager.Shell)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unable to calculate pin cost %s", err.Error()),
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"total_cost_usd": totalCost,
	})
}

func ConfirmPayment(c *gin.Context) {
	paymentID := c.Param("paymentID")
	txHash, present := c.GetPostForm("tx_hash")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tx_hash post form param not present",
		})
		return
	}

	dbUser := c.MustGet("db_user").(string)
	dbPass := c.MustGet("db_pass").(string)
	dbURL := c.MustGet("db_url").(string)
	ethAccount := c.MustGet("eth_account").([2]string) // 0 = key, 1 = pass

	db, err := database.OpenDBConnection(dbPass, dbURL, dbUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to open db connection",
		})
		return
	}
	paymentModelManager := models.NewPaymentManager(db)
	payment := paymentModelManager.FindPaymentByPaymentID(paymentID)
	if payment.CreatedAt == utils.NilTime {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "supplied payment ID does not exist in database",
		})
		return
	}
	if payment.Paid == true {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "supplied payment ID has already been paid for",
		})
		return
	}
	fmt.Println(txHash, ethAccount[0])
	c.JSON(http.StatusOK, gin.H{
		"status": "payment confirmed",
	})
}
