package v2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/ChainRider-Go/dash"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	greq "github.com/RTradeLtd/grpc/temporal/request"
	"github.com/gin-gonic/gin"
)

// ConfirmPayment is used to confirm a payment after sending it.
// By giving Temporal the TxHash, we can then validate that hte payment
// was made, and validated by the appropriate blockchain.
func (api *API) ConfirmPayment(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "payment_number", "tx_hash")
	if len(forms) == 0 {
		return
	}
	paymentNumberInt, err := strconv.ParseInt(forms["payment_number"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	if _, err := api.pm.FindPaymentByNumber(username, paymentNumberInt); err != nil {
		api.LogError(err, eh.PaymentSearchError)(c, http.StatusBadRequest)
		return
	}
	payment, err := api.pm.UpdatePaymentTxHash(username, forms["tx_hash"], paymentNumberInt)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	paymentConfirmation := queue.PaymentConfirmation{
		UserName:      username,
		PaymentNumber: paymentNumberInt,
	}
	if err = api.queues.payConfirm.PublishMessage(paymentConfirmation); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": payment})
}

// RequestSignedPaymentMessage is used to get a signed message from the GRPC API Payments Server
func (api *API) RequestSignedPaymentMessage(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "payment_type", "sender_address", "credit_value")
	if len(forms) == 0 {
		return
	}
	var (
		paymentType string
		method      uint64
	)
	switch forms["payment_type"] {
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
	// get the current value of a single (ie, 1.0 eth) unit of currency of the given payment type
	usdValueFloat, err := api.getUSDValue(paymentType)
	if err != nil {
		api.LogError(err, eh.CmcCheckError)(c, http.StatusBadRequest)
		return
	}
	// get the number of the current payment we are processing
	paymentNumber, err := api.pm.GetLatestPaymentNumber(username)
	if err != nil && err != gorm.ErrRecordNotFound {
		api.LogError(err, eh.PaymentSearchError)(c, http.StatusBadRequest)
		return
	}
	// convert the credits the user wants to buy from string to float
	creditValueFloat, err := strconv.ParseFloat(forms["credit_value"], 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// calculate how much of the given currency we  need to charge them
	chargeAmountFloat := creditValueFloat / usdValueFloat
	// convert the float to a big int, as whenever we are processing uint256 in our smart contracts, this is the equivalent of a big.Int in golang
	chargeAmountBig := utils.FloatToBigInt(chargeAmountFloat)
	// format the big int, as a string
	chargeAmountString := chargeAmountBig.String()
	// do some formatting
	numberString := strconv.FormatInt(paymentNumber, 10)
	methodString := strconv.FormatUint(method, 10)
	// the following pieces of information are used to construct a hash which we then sign
	// using this signed hash, we can then present it to the smart contract, along with the data needed to reconstruct the hash
	// by presenting this information to the smart contract via a transaction sent by the senderAddress, we can validate our payment
	// on-chain, in a trustless manner ensuring transfer of payment and validation of payment within a single smart contract call.
	signRequest := greq.SignRequest{
		// the address that will be sending the transactoin
		Address: forms["sender_address"],
		// the method of the payment
		Method: methodString,
		// the number of the current payment
		Number: numberString,
		// the amount we are charging them
		ChargeAmount: chargeAmountString,
	}
	// send a call to the signer service, which will take the data, hash it, and sign it
	// using the returned values, we have the information needed to send a call to the smart contract
	resp, err := api.signer.GetSignedMessage(context.Background(), &signRequest)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	paymentNumberString := fmt.Sprintf("%s-%s", username, strconv.FormatInt(paymentNumber, 10))
	if _, err = api.pm.NewPayment(
		paymentNumber,
		paymentNumberString,
		paymentNumberString,
		chargeAmountFloat,
		creditValueFloat,
		"ethereum",
		paymentType,
		username,
	); err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	vUint, err := strconv.ParseUint(resp.GetV(), 10, 64)
	if err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	formattedH := fmt.Sprintf("0x%s", resp.GetH())
	formattedR := fmt.Sprintf("0x%s", resp.GetR())
	formattedS := fmt.Sprintf("0x%s", resp.GetS())
	response := gin.H{
		"charge_amount_big": chargeAmountString,
		"method":            uint8(method),
		"payment_number":    paymentNumber,
		"prefixed":          true,
		"v":                 uint8(vUint),
		"formatted": gin.H{
			"h": formattedH,
			"r": formattedR,
			"s": formattedS,
		},
	}
	Respond(c, http.StatusOK, gin.H{"response": response})
}

// CreateDashPayment is used to create a dash payment via chainrider
func (api *API) CreateDashPayment(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "credit_value")
	if len(forms) == 0 {
		return
	}
	usdValueFloat, err := api.getUSDValue("dash")
	if err != nil {
		Fail(c, err)
		return
	}
	creditValueFloat, err := strconv.ParseFloat(forms["credit_value"], 64)
	if err != nil {
		Fail(c, err)
		return
	}
	chargeAmountFloat := creditValueFloat / usdValueFloat
	paymentNumber, err := api.pm.GetLatestPaymentNumber(username)
	if err != nil {
		api.LogError(err, eh.PaymentSearchError)(c, http.StatusBadRequest)
		return
	}
	// as we require tx hashes be unique in the database
	// we need to create a fake, but also unique value as a temporary place holder
	fakeTxHash := fmt.Sprintf("%s-%v", username, paymentNumber)
	// dash is only up to 8 decimals, so we must parse accordingly
	chargeAmountParsed := fmt.Sprintf("%.8f", chargeAmountFloat)
	chargeAmountFloat, err = strconv.ParseFloat(chargeAmountParsed, 64)
	if err != nil {
		Fail(c, err)
	}
	response, err := api.dc.CreatePaymentForward(
		&dash.PaymentForwardOpts{
			DestinationAddress: api.cfg.Wallets.DASH,
		},
	)
	if err != nil {
		api.LogError(err, eh.ChainRiderAPICallError)(c, http.StatusBadRequest)
		return
	}
	if response.Error != "" {
		api.LogError(errors.New(response.Error), eh.ChainRiderAPICallError)(c, http.StatusBadRequest)
		return
	}
	if _, err = api.pm.NewPayment(
		paymentNumber,
		response.PaymentAddress,
		fakeTxHash,
		creditValueFloat,
		chargeAmountFloat,
		"dash",
		"dash",
		username,
	); err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	confirmation := &queue.DashPaymenConfirmation{
		UserName:         username,
		PaymentForwardID: response.PaymentForwardID,
		PaymentNumber:    paymentNumber,
	}
	if err = api.queues.dash.PublishMessage(confirmation); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	type pay struct {
		PaymentNumber    int64
		ChargeAmount     float64
		Blockchain       string
		Status           string
		Network          string
		DepositAddress   string
		PaymentForwardID string
	}
	// calculate the mining fee required by chainrider to forward the payment
	miningFeeDash := dash.DuffsToDash(float64(int64(response.MiningFeeDuffs)))
	// update the charge amount with the mining fee
	chargeAmountFloat = chargeAmountFloat + miningFeeDash
	p := pay{
		PaymentNumber: paymentNumber,
		ChargeAmount:  chargeAmountFloat,
		Blockchain:    "dash",
		Status:        "please send exactly the charge amount. The mining fee required by the chainrider payment forward api call is incldued in the charge amount",
		//TODO: change to main before production release
		Network:          "testnet",
		DepositAddress:   response.PaymentAddress,
		PaymentForwardID: response.PaymentForwardID,
	}
	Respond(c, http.StatusOK, gin.H{"response": p})
}

// GetDepositAddress is used to get a deposit address for a user
func (api *API) GetDepositAddress(c *gin.Context) {
	paymentType := c.Param("type")
	address, err := api.getDepositAddress(paymentType)
	if err != nil {
		api.LogError(err, eh.InvalidPaymentTypeError)(c, http.StatusBadRequest)
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
	case "dash":
		return utils.RetrieveUsdPrice("dash")
	case "btc":
		return utils.RetrieveUsdPrice("bitcoin")
	case "ltc":
		return utils.RetrieveUsdPrice("litecoin")
	case "rtc":
		return RtcCostUsd, nil
	}
	return 0, errors.New(eh.InvalidPaymentTypeError)
}
