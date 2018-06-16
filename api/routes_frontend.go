package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/payment_server"
	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/utils"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/gin-gonic/gin"
)

/*
Contains routes used for frontend operation
*/

// SubmitPinPaymentRequest is used to submit a payment registration
// request by an authenticated user
func SubmitPinPaymentRequest(c *gin.Context) {
	var method uint8
	contextCopy := c.Copy()
	uploaderAddress := GetAuthenticatedUserFromContext(contextCopy)

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
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
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
				"error": fmt.Sprintf("error %s", err.Error()),
			})
			return
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
		"status":         "payment registration request sent",
		"cost_usd":       pinCostUsd,
		"currency_cost":  cost,
		"payment_method": paymentMethod,
	})
}

// CalculatePinCost is used to calculate the cost of pinning something to temporal
func CalculatePinCost(c *gin.Context) {
	hash := c.Param("hash")
	holdTime := c.Param("holdtime")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
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
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total_cost_usd": totalCost,
	})
}

// ConfirmPayment is used to confirm a payment, initiating an upload to temporal
func ConfirmPayment(c *gin.Context) {
	paymentID := c.Param("paymentID")
	txHash, present := c.GetPostForm("tx_hash")
	if !present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tx_hash post form param not present",
		})
		return
	}

	//TODO:  check to make sure the payment id belongs to the authenticated user
	uploaderAddress := GetAuthenticatedUserFromContext(c)

	ethAccount := c.MustGet("eth_account").([2]string) // 0 = key, 1 = pass

	db := c.MustGet("db").(*gorm.DB)

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

	pm, err := payment_server.NewPaymentManager(false, ethAccount[0], ethAccount[1], db)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to establish connection to payment contract",
		})
		return
	}
	var mined bool
	confirmations := 12
	currentConfirmations := 0
	// we will cache the value in case we get an intermittent RPC error
	currentBlockCache, err := pm.EthRPC.EthBlockNumber()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unable to read block number %s", err.Error()),
		})
		return
	}

	for {
		if mined {
			break
		}
		receipt, err := pm.EthRPC.EthGetTransactionReceipt(txHash)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("error processing tx receipt %s", err.Error()),
			})
			return
		}

		if receipt.CumulativeGasUsed > 0 && receipt.BlockNumber > 0 {
			blockMinedAt := receipt.BlockNumber
			currentBlock, err := pm.EthRPC.EthBlockNumber()
			// if there was an erro reading the block number, lets use our cached value
			if err != nil {
				currentBlock = currentBlockCache
			}
			currentConfirmations = currentBlock - blockMinedAt
			if currentConfirmations >= confirmations {
				mined = true
			}
		}
	}

	// payment has been confirmed so lets process this
	mqURL := c.MustGet("mq_conn_url").(string)
	qm, err := queue.Initialize(queue.PaymentReceivedQueue, mqURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("unable to open queue manager %s", err.Error()),
		})
		return
	}

	paymentReceived := queue.PaymentReceived{
		UploaderAddress: uploaderAddress,
		PaymentID:       paymentID,
	}

	// this will trigger a message to rabbitmq, prompting pinning of the content to temporal
	err = qm.PublishMessage(paymentReceived)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("error publishing message to queue %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "payment confirmed",
	})
}
