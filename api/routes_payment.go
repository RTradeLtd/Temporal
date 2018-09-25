package api

import (
	"errors"
	"net/http"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/gin-gonic/gin"
)

// CreatePayment is used to create a payment
func (api *API) CreatePayment(c *gin.Context) {
	paymentType := c.Param("type")
	// place holder
	// call CMC with currency to fetch USD value
	usdValue := "temporary"
	// place holder
	// generate unique deposit address
	depositAddress := "temporary"
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
	username := GetAuthenticatedUserFromContext(c)
	pm := models.NewPaymentManager(api.DBM.DB)
	if _, err := pm.NewPayment(depositAddress, txHash, usdValue, blockchain, paymentType, username); err != nil {
		api.LogError(err, PaymentCreationError)
		FailOnError(c, err)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "payment created"})
}

// GetDepositAddress is used to get a deposit address for a user
func (api *API) GetDepositAddress(c *gin.Context) {
	paymentType := c.Param("type")
	username := GetAuthenticatedUserFromContext(c)
	var (
		err     error
		address string
		um      = models.NewUserManager(api.DBM.DB)
	)
	switch paymentType {
	case "ETH", "RTC":
		address, err = um.FindEthAddressByUserName(username)
	default:
		err = errors.New(InvalidPaymentTypeError)
		api.LogError(err, InvalidPaymentTypeError)
		return
	}
	if err != nil {
		api.LogError(err, EthAddressSearchError)
		FailOnError(c, err)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": address})
}
