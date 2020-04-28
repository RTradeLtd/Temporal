package v2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gcash/bchutil"

	"github.com/stripe/stripe-go/charge"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/jinzhu/gorm"
	"github.com/stripe/stripe-go"

	"github.com/RTradeLtd/ChainRider-Go/dash"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	greq "github.com/RTradeLtd/grpc/pay/request"
	pbBchWallet "github.com/gcash/bchwallet/rpc/walletrpc"

	"github.com/gin-gonic/gin"
)

// ConfirmETHPayment is used to confirm an ethereum based payment
func (api *API) ConfirmETHPayment(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract post forms
	forms, missingField := api.extractPostForms(c, "payment_number", "tx_hash")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// parse payment number
	paymentNumberInt, err := strconv.ParseInt(forms["payment_number"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// check to see if this payment is already registered
	payment, err := api.pm.FindPaymentByNumber(username, paymentNumberInt)
	if err != nil {
		api.LogError(c, err, eh.PaymentSearchError)(http.StatusBadRequest)
		return
	}
	// validate that the payment is for the appropriate blockchain
	if payment.Blockchain != "ethereum" {
		Fail(c, errors.New("payment you are trying to confirm is not for the ethereum blockchain"))
		return
	}
	// this is used to prevent people from abusing the payment system, and getting
	// a single payment to be processed multiple times without having to send additional funds
	if payment.TxHash[0:2] == "0x" {
		Fail(c, errors.New("payment is already being processed, if your payment hasn't been confirmed after 90 minutes please contact support@rtradetechnologies.com"))
		return
	}
	// update payment with the new tx hash
	if _, err = api.pm.UpdatePaymentTxHash(username, forms["tx_hash"], paymentNumberInt); err != nil {
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
	forms, missingField := api.extractPostForms(c, "payment_type", "sender_address", "credit_value")
	if missingField != "" {
		FailWithMissingField(c, missingField)
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

func (api *API) createBchPayment(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms, missingField := api.extractPostForms(c, "credit_value")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	usdValueFloat, err := api.getUSDValue("bch")
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
	/*
			addrReq, err := api.bchWallet.CurrentAddress(context.Background(), &pbBchWallet.CurrentAddressRequest{Account: 0})
			if err != nil {
				api.LogError(c, err, "failed to get bch deposit address", http.StatusInternalServerError)
				return
			}
		one thing to note about the below solution, this is technically only usable with a fully synced blockchain
		so it may/may not cause issues in the future. If this does cause issues, then we will need to disable
		this section, and enable the previously commented section above
		this is temporary fix until the underlying issue is fixed
		https://github.com/gcash/bchwallet/issues/41
	*/
	addrReq, err := api.bchWallet.NextAddress(context.Background(), &pbBchWallet.NextAddressRequest{Account: 0})
	if err != nil {
		api.LogError(c, err, "failed to get bch deposit address", http.StatusInternalServerError)
		return
	}
	bchChargeAmount, err := bchutil.NewAmount(chargeAmountFloat)
	if err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	// format a unique payment number to take the place of deposit address and tx hash temporarily
	paymentNumberString := fmt.Sprintf("%s-%s", username, strconv.FormatInt(paymentNumber, 10))
	if _, err := api.pm.NewPayment(
		paymentNumber,
		addrReq.GetAddress(),
		// temporary fake tx hash
		// will get updated when
		// confirming the payment
		paymentNumberString,
		creditValueFloat,
		bchChargeAmount.ToBCH(),
		"bitcoin-cash",
		"bch",
		username,
	); err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	response := gin.H{
		"deposit_address": addrReq.GetAddress(),
		"charge_amount":   bchChargeAmount.ToBCH(),
		"payment_number":  paymentNumber,
	}
	Respond(c, http.StatusOK, gin.H{"response": response})
}

func (api *API) confirmBchPayment(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms, missingField := api.extractPostForms(c, "payment_number", "tx_hash")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// parse payment number
	paymentNumberInt, err := strconv.ParseInt(forms["payment_number"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// check to see if this payment is already registered
	payment, err := api.pm.FindPaymentByNumber(username, paymentNumberInt)
	if err != nil {
		api.LogError(c, err, eh.PaymentSearchError)(http.StatusBadRequest)
		return
	}
	if payment.Blockchain != "bitcoin-cash" {
		Fail(c, errors.New("payment you are trying to confirm is not for the bitcoin-cash blockchain"))
		return
	}
	if payment.Confirmed {
		Fail(c, errors.New("payment is already confirmed"))
		return
	}
	// when a payment is first created, we insert temporary data into the TxHash field
	// this field only gets updated when confirming a payment. In order to prevent
	// abuse and multiple transactions being submitted for the same payment
	// before it is confirmed, ensure that we only allow processing if the temporary
	// TxHash is present
	if payment.TxHash != fmt.Sprintf("%s-%s", username, strconv.FormatInt(payment.Number, 10)) {
		Fail(c, errors.New("payment is already being processed"))
		return
	}
	if _, err := api.pm.UpdatePaymentTxHash(username, forms["tx_hash"], paymentNumberInt); err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	confirmation := queue.BchPaymentConfirmation{
		UserName:      username,
		PaymentNumber: paymentNumberInt,
	}
	if err := api.queues.bch.PublishMessage(confirmation); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": confirmation})
}

// CreateDashPayment is used to create a dash payment via chainrider
func (api *API) CreateDashPayment(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms, missingField := api.extractPostForms(c, "credit_value")
	if missingField != "" {
		FailWithMissingField(c, missingField)
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
		Network:          "main",
		DepositAddress:   response.PaymentAddress,
		PaymentForwardID: response.PaymentForwardID,
	}
	Respond(c, http.StatusOK, gin.H{"response": p})
}

// stripChargeTemporalX enables purchasing the license for TemporalX
func (api *API) stripeChargeTemporalX(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms, missingField := api.extractPostForms(c, "stripe_token", "stripe_email")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// set the secret key after input validation
	stripe.Key = api.cfg.Stripe.SecretKey
	// set the source for the charge
	// in this case it is a tokenized version of the credit card
	source, err := stripe.SourceParamsFor(forms["stripe_token"])
	if err != nil {
		Fail(c, err)
		return
	}
	// initialize credit card charge parameters
	ch, err := charge.New(&stripe.ChargeParams{
		Amount:      stripe.Int64(335), // equivalent to purchase cose + GST&PST provincial taxes
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		Description: stripe.String("temporalx 3-node license purchase"),
		// StatementDescriptor is what appears in their credit card billing report
		StatementDescriptor: stripe.String("credit purchase"),
		// email the receipt goes to
		ReceiptEmail: stripe.String(forms["stripe_email"]),
		Source:       source,
		Params: stripe.Params{
			Metadata: map[string]string{
				"order_type": "temporal.credits",
			},
		},
	})
	if err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	api.l.Infow("payment complete", "payment.method", "stripe", "user", username, "charge", ch)
	if err := api.queues.email.PublishMessage(queue.EmailSend{
		Subject:     "TemporalX License Purchase",
		Content:     username + " has purchased a 3 node license for TemporalX",
		ContentType: "text/html",
		UserNames:   []string{username},
		Emails:      []string{"admin@rtradetechnologies.com"},
	}); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "temporalx license purchase succcessful, details will be sent to email on file within 24 hours"})
}

func (api *API) stripeCharge(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms, missingField := api.extractPostForms(c, "stripe_token", "stripe_email", "value_in_cents")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	valueInCentsInt, err := strconv.ParseInt(forms["value_in_cents"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	valueInCentsFloat, err := strconv.ParseFloat(forms["value_in_cents"], 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// set the secret key after input validation
	stripe.Key = api.cfg.Stripe.SecretKey
	// set the source for the charge
	// in this case it is a tokenized version of the credit card
	source, err := stripe.SourceParamsFor(forms["stripe_token"])
	if err != nil {
		Fail(c, err)
		return
	}
	// initialize credit card charge parameters
	ch, err := charge.New(&stripe.ChargeParams{
		Amount:      stripe.Int64(valueInCentsInt),
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		Description: stripe.String("temporal credit purchase"),
		// StatementDescriptor is what appears in their credit card billing report
		StatementDescriptor: stripe.String("credit purchase"),
		// email the receipt goes to
		ReceiptEmail: stripe.String(forms["stripe_email"]),
		Source:       source,
		Params: stripe.Params{
			Metadata: map[string]string{
				"order_type": "temporal.credits",
			},
		},
	})
	if err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	api.l.Infow("payment complete", "payment.method", "stripe", "user", username, "charge", ch)
	// add credits to use
	if _, err := api.um.AddCredits(username, valueInCentsFloat/100); err != nil {
		api.LogError(c, err, "failed to grant credits")(http.StatusInternalServerError)
		return
	}
	api.l.Infow("credits granted", "payment.method", "stripe", "credit.amount", valueInCentsFloat/100)
	Respond(c, http.StatusOK, gin.H{"response": "stripe credit purchase successful"})
}

// GetPaymentStatus is used to retrieve whether or not a payment is confirmed
func (api *API) getPaymentStatus(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	number, err := strconv.ParseInt(c.Param("number"), 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	payment, err := api.pm.FindPaymentByNumber(username, number)
	if err != nil {
		api.LogError(c, err, eh.PaymentSearchError)(http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": payment.Confirmed})
}

// GetUSDValue is used to retrieve the usd value of a given payment type
func (api *API) getUSDValue(paymentType string) (float64, error) {
	var (
		cost float64
		err  error
	)
	switch paymentType {
	case "eth":
		cost, err = utils.RetrieveUsdPrice("ethereum", api.getCMCKey())
	case "xmr":
		cost, err = utils.RetrieveUsdPrice("monero", api.getCMCKey())
	case "dash":
		cost, err = utils.RetrieveUsdPrice("dash", api.getCMCKey())
	case "btc":
		cost, err = utils.RetrieveUsdPrice("bitcoin", api.getCMCKey())
	case "bch":
		cost, err = utils.RetrieveUsdPrice("bitcoin-cash", api.getCMCKey())
	case "ltc":
		cost, err = utils.RetrieveUsdPrice("litecoin", api.getCMCKey())
	case "rtc":
		cost, err = RtcCostUsd, nil
	default:
		return 0, errors.New(eh.InvalidPaymentTypeError)
	}
	// if we have an error, and the cost is 0 return the error
	// otherwise if we have an error, and a non-zero cost
	// then we should return the most recent price we have
	if err != nil && cost == 0 {
		return 0, err
	}
	return cost, nil
}
