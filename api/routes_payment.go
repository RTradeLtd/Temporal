package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/Temporal/gapi"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	greq "github.com/RTradeLtd/gapimit/request"
	"github.com/gin-gonic/gin"
)

// ConfirmPayment is used to confirm a payment after sending it.
// By giving Temporal the TxHash, we can then validate that hte payment
// was made, and validated by the appropriate blockchain.
func (api *API) ConfirmPayment(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	paymentNumber, exists := c.GetPostForm("payment_number")
	if !exists {
		FailWithMissingField(c, "payment_number")
		return
	}
	txHash, exists := c.GetPostForm("tx_hash")
	if !exists {
		FailWithMissingField(c, "tx_hash")
		return
	}
	paymentNumberInt, err := strconv.ParseInt(paymentNumber, 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	pm := models.NewPaymentManager(api.dbm.DB)
	if _, err := pm.FindPaymentByNumber(username, paymentNumberInt); err != nil {
		api.LogError(err, PaymentSearchError)(c, http.StatusBadRequest)
		return
	}
	payment, err := pm.UpdatePaymentTxHash(username, txHash, paymentNumberInt)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	paymentConfirmation := queue.PaymentConfirmation{
		UserName:      username,
		PaymentNumber: paymentNumberInt,
	}
	qm, err := queue.Initialize(
		queue.PaymentConfirmationQueue,
		api.cfg.RabbitMQ.URL,
		true,
		false,
	)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c, http.StatusBadRequest)
		return
	}
	if err = qm.PublishMessage(paymentConfirmation); err != nil {
		api.LogError(err, QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": payment})
}

// RequestSignedPaymentMessage is used to get a signed message from the GRPC API Payments Server
func (api *API) RequestSignedPaymentMessage(c *gin.Context) {
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
	if err != nil && err != gorm.ErrRecordNotFound {
		api.LogError(err, PaymentSearchError)(c, http.StatusBadRequest)
		return
	}
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
	paymentNumberString := strconv.FormatInt(paymentNumber, 10)

	if _, err = pm.NewPayment(
		paymentNumber,
		paymentNumberString,
		paymentNumberString,
		usdValueFloat,
		"ethereum",
		paymentType,
		username,
	); err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	response := gin.H{
		"charge_amount":  signRequest.ChargeAmount,
		"method":         signRequest.Method,
		"payment_number": paymentNumber,
		"prefixed":       true,
		"sig_parts":      resp,
	}
	Respond(c, http.StatusOK, gin.H{"response": response})
}

// CreatePayment is used to create a payment
func (api *API) CreatePayment(c *gin.Context) {
	Respond(c, http.StatusNotImplemented, gin.H{"response": "this api call isn't fully implemented yet"})
	return
	username := GetAuthenticatedUserFromContext(c)
	paymentType, exists := c.GetPostForm("payment_type")
	if !exists {
		FailWithMissingField(c, "payment_type")
		return
	}
	switch paymentType {
	case "eth", "rtc":
		err := errors.New("for 'rtc' and 'eth' payments please use the request route")
		Fail(c, err, http.StatusBadRequest)
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
