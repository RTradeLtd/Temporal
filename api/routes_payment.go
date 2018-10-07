package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/gapi"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	greq "github.com/RTradeLtd/gapimit/request"
	"github.com/gin-gonic/gin"
)

// GetSignedMessage is used to get a signed message from the GRPC API Payments Server
func (api *API) GetSignedMessage(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	paymentType, exists := c.GetPostForm("payment_type")
	if !exists {
		FailWithMissingField(c, "payment_type")
		return
	}
	var method uint64
	switch paymentType {
	case "0":
		paymentType = "rtc"
		method = 0
	case "1":
		paymentType = "eth"
		method = 1
	default:
		Fail(c, errors.New("payment_type must be '0 (rtc)' or '1 (eth)'"))
		return
	}
	senderAddress, exists := c.GetPostForm("sender_address")
	if !exists {
		FailWithMissingField(c, "sender_address")
		return
	}
	creditValue, exists := c.GetPostForm("credit_value")
	if !exists {
		FailWithMissingField(c, "credit_value")
		return
	}
	usdValueFloat, err := api.getUSDValue(paymentType)
	if err != nil {
		api.LogError(err, CmcCheckError)(c, http.StatusBadRequest)
		return
	}
	pm := models.NewPaymentManager(api.dbm.DB)
	paymentNumber, err := pm.GetLatestPaymentNumber(username)
	if err != nil {
		api.LogError(err, PaymentSearchError)(c, http.StatusBadRequest)
		return
	}
	paymentNumber++
	creditValueFloat, err := strconv.ParseFloat(creditValue, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	chargeAmountFloat := creditValueFloat / usdValueFloat
	chargeAmountString := strconv.FormatFloat(chargeAmountFloat, 'f', 18, 64)
	numberString := strconv.FormatInt(paymentNumber, 10)
	methodString := strconv.FormatUint(method, 10)
	signRequest := greq.SignRequest{
		Address:      senderAddress,
		Method:       methodString,
		Number:       numberString,
		ChargeAmount: chargeAmountString,
	}
	gc, err := gapi.NewGAPIClient(api.cfg, true)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	defer gc.GC.Close()
	resp, err := gc.GetSignedMessage(context.Background(), &signRequest)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": resp})
}

// CreatePayment is used to create a payment
func (api *API) CreatePayment(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	paymentType, exists := c.GetPostForm("payment_type")
	if !exists {
		FailWithMissingField(c, "payment_type")
		return
	}
	usdValue, err := api.getUSDValue(paymentType)
	if err != nil {
		api.LogError(err, CmcCheckError)(c, http.StatusBadRequest)
		return
	}
	depositAddress, err := api.getDepositAddress(paymentType)
	if err != nil {
		api.LogError(err, DepositAddressCheckError)(c, http.StatusBadRequest)
		return
	}
	txHash, exists := c.GetPostForm("tx_hash")
	if !exists {
		FailWithMissingField(c, "tx_hash")
		return
	}
	blockchain, exists := c.GetPostForm("blockchain")
	if !exists {
		FailWithMissingField(c, "blockchain")
		return
	}
	if check := api.validateBlockchain(blockchain); !check {
		api.LogError(err, InvalidPaymentBlockchainError)(c, http.StatusBadRequest)
		return
	}
	pm := models.NewPaymentManager(api.dbm.DB)
	latestPaymentNumber, err := pm.GetLatestPaymentNumber(username)
	if err != nil {
		api.LogError(err, PaymentSearchError)(c, http.StatusBadRequest)
		return
	}
	if _, err := pm.NewPayment(latestPaymentNumber, depositAddress, txHash, usdValue, blockchain, paymentType, username); err != nil {
		api.LogError(err, PaymentCreationError)(c, http.StatusBadRequest)
		return
	}
	pc := queue.PaymentCreation{
		TxHash:     txHash,
		Blockchain: blockchain,
		UserName:   username,
	}
	mqURL := api.cfg.RabbitMQ.URL
	qm, err := queue.Initialize(queue.PaymentCreationQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c, http.StatusBadRequest)
		return
	}
	if err = qm.PublishMessage(pc); err != nil {
		api.LogError(err, QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "payment created"})
}

// GetDepositAddress is used to get a deposit address for a user
func (api *API) GetDepositAddress(c *gin.Context) {
	paymentType := c.Param("type")
	address, err := api.getDepositAddress(paymentType)
	if err != nil {
		api.LogError(err, InvalidPaymentTypeError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": address})
}

// GetUSDValue is used to retrieve the usd value of a given payment type
func (api *API) getUSDValue(paymentType string) (float64, error) {
	switch paymentType {
	case "eth":
		return utils.RetrieveUsdPrice("ethereum")
	case "xmr":
		return utils.RetrieveUsdPrice("monero")
	case "btc":
		return utils.RetrieveUsdPrice("bitcoin")
	case "ltc":
		return utils.RetrieveUsdPrice("litecoin")
	case "rtc":
		return RtcCostUsd, nil
	}
	return 0, errors.New(InvalidPaymentTypeError)
}
