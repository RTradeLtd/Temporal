package api

import (
	"errors"
	"net/http"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
)

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
	if _, err := pm.NewPayment(depositAddress, txHash, usdValue, blockchain, paymentType, username); err != nil {
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
