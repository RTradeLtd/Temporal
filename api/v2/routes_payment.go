package v2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/gorm"

	"github.com/RTradeLtd/ChainRider-Go/dash"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	greq "github.com/RTradeLtd/grpc/pay/request"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"github.com/stripe/stripe-go/customer"
)

// ConfirmPayment is used to confirm a payment after sending it.
// By giving Temporal the TxHash, we can then validate that hte payment
// was made, and validated by the appropriate blockchain.
func (api *API) ConfirmPayment(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract post forms
	forms := api.extractPostForms(c, "payment_number", "tx_hash")
	if len(forms) == 0 {
		return
	}
	// parse payment number
	paymentNumberInt, err := strconv.ParseInt(forms["payment_number"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// check to see if this payment is already registered
	if _, err := api.pm.FindPaymentByNumber(username, paymentNumberInt); err != nil {
		api.LogError(c, err, eh.PaymentSearchError)(http.StatusBadRequest)
		return
	}
	// update payment with the new tx hash
	payment, err := api.pm.UpdatePaymentTxHash(username, forms["tx_hash"], paymentNumberInt)
	if err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	// create payment confirmation message
	paymentConfirmation := queue.EthPaymentConfirmation{
		UserName:      username,
		PaymentNumber: paymentNumberInt,
	}
	// send message for processing
	if err = api.queues.eth.PublishMessage(paymentConfirmation); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": payment})
}

// RequestSignedPaymentMessage is used to get a signed message from the GRPC API Payments Server
// this is currently used for ETH+RTC smart-contract facilitated payments
func (api *API) RequestSignedPaymentMessage(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract post forms
	forms := api.extractPostForms(c, "payment_type", "sender_address", "credit_value")
	if len(forms) == 0 {
		return
	}
	var (
		paymentType string
		method      uint64
	)
	// ensure it is a valid payment type
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
		api.LogError(c, err, eh.CmcCheckError)(http.StatusBadRequest)
		return
	}
	// get the number of the current payment we are processing
	paymentNumber, err := api.pm.GetLatestPaymentNumber(username)
	if err != nil && err != gorm.ErrRecordNotFound {
		api.LogError(c, err, eh.PaymentSearchError)(http.StatusBadRequest)
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
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	// format a unique payment number to take the place of deposit address and tx hash temporarily
	paymentNumberString := fmt.Sprintf("%s-%s", username, strconv.FormatInt(paymentNumber, 10))
	if _, err = api.pm.NewPayment(
		paymentNumber,
		paymentNumberString,
		paymentNumberString,
		creditValueFloat,
		chargeAmountFloat,
		"ethereum",
		paymentType,
		username,
	); err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	// parse the v parameter into uint type
	vUint, err := strconv.ParseUint(resp.GetV(), 10, 64)
	if err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	// format signed message components to send to be used
	// for submission to contracts
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
	// return
	Respond(c, http.StatusOK, gin.H{"response": response})
}

// CreateDashPayment is used to create a dash payment via chainrider
func (api *API) CreateDashPayment(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
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
		api.LogError(c, err, eh.PaymentSearchError)(http.StatusBadRequest)
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
		api.LogError(c, err, eh.ChainRiderAPICallError, "wallet_address", api.cfg.Wallets.DASH)(http.StatusBadRequest)
		return
	}
	if response.Error != "" {
		api.LogError(c, errors.New(response.Error), eh.ChainRiderAPICallError, "wallet_address", api.cfg.Wallets.DASH)(http.StatusBadRequest)
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
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	confirmation := &queue.DashPaymenConfirmation{
		UserName:         username,
		PaymentForwardID: response.PaymentForwardID,
		PaymentNumber:    paymentNumber,
	}
	if err = api.queues.dash.PublishMessage(confirmation); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
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
		api.LogError(c, err, eh.InvalidPaymentTypeError)(http.StatusBadRequest)
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

// stripeDisplay is used to display the strip checkout page
func (api *API) stripeDisplay(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, "no api token provided")(http.StatusBadRequest)
		return
	}
	// find email related to user
	user, err := api.um.FindByUserName(username)
	if err != nil {
		Fail(c, err, http.StatusBadRequest)
		return
	}
	// render the actuall checkout box
	c.HTML(http.StatusOK, "stripe.html", gin.H{
		"Key":         api.cfg.Stripe.PublishableKey,
		"title":       "Temporal Credit Purchase",
		"description": "Purchase credits using stripe",
		"amount":      c.Param("cents"),
		"email":       user.EmailAddress,
	})
}

// stripeCharge is used to place the actual credit card charge
func (api *API) stripeCharge(c *gin.Context) {
	// get the user account associated with the email
	user, err := api.um.FindByEmail(c.PostForm("stripeEmail"))
	if err != nil {
		api.LogError(c, err, "provided email account does not exist, use temporal email")(http.StatusBadRequest)
		return
	}
	// setup stripe instance
	stripe.Key = api.cfg.Stripe.SecretKey
	// setup customer objecter
	customerParams := &stripe.CustomerParams{
		Email: stripe.String(c.PostForm("stripeEmail")),
	}
	// update customer object
	customerParams.SetSource(c.PostForm("stripeToken"))
	// instantiate customer
	newCustomer, err := customer.New(customerParams)
	if err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	// extract the amount of cents to charge them
	creditValueCentsString := c.PostForm("provided_amount")
	// format amount of cents from string to int
	creditValueCentsInt, err := strconv.ParseInt(creditValueCentsString, 10, 64)
	if err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	// instantiate charge parameters
	chargeParams := &stripe.ChargeParams{
		Amount:      stripe.Int64(creditValueCentsInt), // this is the amount of cents to charge
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		Description: stripe.String("Temporal Credit Purchase"),
		Customer:    stripe.String(newCustomer.ID),
	}
	// place the actual charge
	if _, err := charge.New(chargeParams); err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	// add the credits
	if _, err := api.um.AddCredits(user.UserName, float64(creditValueCentsInt)/100); err != nil {
		api.LogError(c, err, "failed to grant credits")(http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "charge successfully placed"})
}
