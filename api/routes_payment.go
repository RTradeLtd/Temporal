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
		FailNoExistPostForm(c, "payment_type")
		return
	}
	usdValue, err := api.getUSDValue(paymentType)
	if err != nil {
		FailOnError(c, err)
		return
	}
	depositAddress, err := api.getDepositAddress(paymentType)
	if err != nil {
		FailOnError(c, err)
		return
	}
	txHash, exists := c.GetPostForm("tx_hash")
	if !exists {
		FailNoExistPostForm(c, "tx_hash")
		return
	}
	blockchain, exists := c.GetPostForm("blockchain")
	if !exists {
		FailNoExistPostForm(c, "blockchain")
		return
	}
	if check := api.validateBlockchain(blockchain); !check {
		FailOnError(c, errors.New(InvalidPaymentBlockchainError))
		return
	}
	pm := models.NewPaymentManager(api.DBM.DB)
	if _, err := pm.NewPayment(depositAddress, txHash, usdValue, blockchain, paymentType, username); err != nil {
		api.LogError(err, PaymentCreationError)
		FailOnError(c, err)
		return
	}
	pc := queue.PaymentCreation{
		TxHash:     txHash,
		Blockchain: blockchain,
		UserName:   username,
	}
	mqURL := api.TConfig.RabbitMQ.URL
	qm, err := queue.Initialize(queue.PaymentCreationQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)
		FailOnError(c, err)
		return
	}
	if err = qm.PublishMessage(pc); err != nil {
		api.LogError(err, QueuePublishError)
		FailOnError(c, err)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "payment created"})
}

// GetDepositAddress is used to get a deposit address for a user
func (api *API) GetDepositAddress(c *gin.Context) {
	paymentType := c.Param("type")
	var address string
	switch paymentType {
	case "eth", "rtc":
		address = "0xc7459562777DDf3A1A7afefBE515E8479Bd3FDBD"
	default:
		err := errors.New(InvalidPaymentTypeError)
		api.LogError(err, InvalidPaymentTypeError)
		FailOnError(c, err)
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
		return 0.125, nil
	}
	return 0, errors.New(InvalidPaymentTypeError)
}
