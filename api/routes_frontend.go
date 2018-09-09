package api

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/signer"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
	"github.com/jinzhu/gorm"
	minio "github.com/minio/minio-go"
	log "github.com/sirupsen/logrus"
)

// CalculateIPFSFileHash is used to calculate the ipfs hash of a file
func (api *API) calculateIPFSFileHash(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	file, err := c.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}
	fh, err := file.Open()
	if err != nil {
		api.LogError(err, FileOpenError)
		FailOnError(c, err)
		return
	}
	hash, err := utils.GenerateIpfsMultiHashForFile(fh)
	if err != nil {
		api.LogError(err, IPFSMultiHashGenerationError)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": api,
		"user":    username,
	}).Info("ipfs file hash calculation requested")

	Respond(c, http.StatusOK, gin.H{"response": hash})
}

// CalculatePinCost is used to calculate the cost of pinning something to temporal
func (api *API) calculatePinCost(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		FailOnError(c, err)
		return
	}
	holdTime := c.Param("holdtime")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		api.LogError(err, IPFSConnectionError)
		FailOnError(c, err)
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}
	totalCost, err := utils.CalculatePinCost(hash, holdTimeInt, manager.Shell)
	if err != nil {
		api.LogError(err, PinCostCalculationError)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("pin cost calculation requested")

	Respond(c, http.StatusOK, gin.H{"response": totalCost})
}

// CalculateFileCost is used to calculate the cost of uploading a file to our system
func (api *API) calculateFileCost(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	file, err := c.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}
	holdTime, exists := c.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("file cost calculation requested")

	cost := utils.CalculateFileCost(holdTimeInt, file.Size)
	Respond(c, http.StatusOK, gin.H{"response": cost})
}

// CreatePinPayment is used to create a signed message for a pin payment
func (api *API) createPinPayment(c *gin.Context) {
	contentHash := c.Param("hash")
	if _, err := gocid.Decode(contentHash); err != nil {
		FailOnError(c, err)
		return
	}
	username := GetAuthenticatedUserFromContext(c)
	holdTime, exists := c.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}
	method, exists := c.GetPostForm("payment_method")
	if !exists {
		FailNoExistPostForm(c, "payment_method")
		return
	}
	methodUint, err := strconv.ParseUint(method, 10, 8)
	if err != nil {
		FailOnError(c, err)
		return
	}
	if methodUint > 1 {
		FailOnError(c, errors.New("payment_method must be 1 or 0"))
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}

	manager, err := rtfs.Initialize("", "")
	if err != nil {
		api.LogError(err, IPFSConnectionError)
		FailOnError(c, err)
		return
	}
	totalCost, err := utils.CalculatePinCost(contentHash, holdTimeInt, manager.Shell)
	if err != nil {
		api.LogError(err, PinCostCalculationError)
		FailOnError(c, err)
		return
	}

	keyFile := api.TConfig.Ethereum.Account.KeyFile
	keyPass := api.TConfig.Ethereum.Account.KeyPass
	ps, err := signer.GeneratePaymentSigner(keyFile, keyPass)
	if err != nil {
		api.LogError(err, PaymentSignerGenerationError)
		FailOnError(c, err)
		return
	}
	um := models.NewUserManager(api.DBM.DB)
	ethAddress, err := um.FindEthAddressByUserName(username)
	if err != nil {
		api.LogError(err, EthAddressSearchError)
		FailOnError(c, err)
		return
	}
	ppm := models.NewPaymentManager(api.DBM.DB)
	var num *big.Int
	num, err = ppm.RetrieveLatestPaymentNumberForUser(username)
	if err != nil && err != gorm.ErrRecordNotFound {
		api.LogError(err, PaymentSearchError)
		FailOnError(c, err)
		return
	}
	if num == nil {
		num = big.NewInt(0)
	}
	num = new(big.Int).Add(num, big.NewInt(1))
	costBig := utils.FloatToBigInt(totalCost)
	// for testing purpose
	addressTyped := common.HexToAddress(ethAddress)

	sm, err := ps.GenerateSignedPaymentMessagePrefixed(addressTyped, uint8(methodUint), num, costBig)
	if err != nil {
		api.LogError(err, PaymentMessageSignError)
		FailOnError(c, err)
		return
	}

	if _, err = ppm.NewPayment(uint8(methodUint), sm.PaymentNumber, sm.ChargeAmount, ethAddress, contentHash, username, "pin", "public", holdTimeInt); err != nil {
		api.LogError(err, PaymentCreationError)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service":        "api",
		"user":           username,
		"payment_number": sm.PaymentNumber.String(),
	}).Info("pin payment request generated")

	Respond(c, http.StatusOK, gin.H{"response": gin.H{
		"h":                    sm.H,
		"v":                    sm.V,
		"r":                    sm.R,
		"s":                    sm.S,
		"eth_address":          sm.Address,
		"charge_amount_in_wei": sm.ChargeAmount,
		"payment_method":       sm.PaymentMethod,
		"payment_number":       sm.PaymentNumber}})
}

// CreateFilePayment is used to create a signed file payment message
func (api *API) createFilePayment(c *gin.Context) {
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}
	holdTimeInMonths, exists := c.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}

	method, exists := c.GetPostForm("payment_method")
	if !exists {
		FailNoExistPostForm(c, "payment_method")
		return
	}
	fileHandler, err := c.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}
	methodUint, err := strconv.ParseUint(method, 10, 8)
	if err != nil {
		FailOnError(c, err)
		return
	}
	if methodUint > 1 {
		FailOnError(c, errors.New("payment_method must be 1 or 0"))
		return
	}

	accessKey := api.TConfig.MINIO.AccessKey
	secretKey := api.TConfig.MINIO.SecretKey
	endpoint := fmt.Sprintf("%s:%s", api.TConfig.MINIO.Connection.IP, api.TConfig.MINIO.Connection.Port)
	keyFile := api.TConfig.Ethereum.Account.KeyFile
	keyPass := api.TConfig.Ethereum.Account.KeyPass
	ps, err := signer.GeneratePaymentSigner(keyFile, keyPass)
	if err != nil {
		api.LogError(err, PaymentSignerGenerationError)
		FailOnError(c, err)
		return
	}
	miniManager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, false)
	if err != nil {
		api.LogError(err, MinioConnectionError)
		FailOnError(c, err)
		return
	}

	fmt.Println("opening file")
	openFile, err := fileHandler.Open()
	if err != nil {
		api.LogError(err, FileOpenError)
		FailOnError(c, err)
		return
	}
	fmt.Println("file opened")
	username := GetAuthenticatedUserFromContext(c)

	holdTimeInMonthsInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}
	cost := utils.CalculateFileCost(holdTimeInMonthsInt, fileHandler.Size)
	costBig := utils.FloatToBigInt(cost)
	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", username, randString)
	fmt.Println("storing file in minio")
	if _, err = miniManager.PutObject(FilesUploadBucket, objectName, openFile, fileHandler.Size, minio.PutObjectOptions{}); err != nil {
		api.LogError(err, MinioPutError)
		FailOnError(c, err)
		return
	}
	fmt.Println("file stored in minio")

	pm := models.NewPaymentManager(api.DBM.DB)
	um := models.NewUserManager(api.DBM.DB)
	ethAddress, err := um.FindEthAddressByUserName(username)
	if err != nil {
		api.LogError(err, EthAddressSearchError)
		FailOnError(c, err)
		return
	}
	var num *big.Int
	num, err = pm.RetrieveLatestPaymentNumberForUser(username)
	if err != nil && err != gorm.ErrRecordNotFound {
		api.LogError(err, PaymentSearchError)
		FailOnError(c, err)
		return
	}
	if num == nil {
		num = big.NewInt(0)

	} else if num.Cmp(big.NewInt(0)) == 1 {
		// this means that the latest payment number is greater than 0
		// indicating a payment has already been made, in which case
		// we will increment the value by 1
		num = new(big.Int).Add(num, big.NewInt(1))
	}
	addressTyped := common.HexToAddress(ethAddress)
	sm, err := ps.GenerateSignedPaymentMessagePrefixed(addressTyped, uint8(methodUint), num, costBig)
	if err != nil {
		api.LogError(err, PaymentMessageSignError)
		FailOnError(c, err)
		return
	}
	if _, err = pm.NewPayment(uint8(methodUint), sm.PaymentNumber, sm.ChargeAmount, ethAddress, objectName, username, "file", networkName, holdTimeInMonthsInt); err != nil {
		api.LogError(err, PaymentCreationError)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service":        "api",
		"user":           username,
		"payment_number": sm.PaymentNumber.String(),
	}).Info("file payment request generated")

	Respond(c, http.StatusOK, gin.H{"response": gin.H{
		"h":                    sm.H,
		"v":                    sm.V,
		"r":                    sm.R,
		"s":                    sm.S,
		"eth_address":          sm.Address,
		"charge_amount_in_wei": sm.ChargeAmount,
		"payment_method":       sm.PaymentMethod,
		"payment_number":       sm.PaymentNumber}})
}

// SubmitPinPaymentConfirmation is used to submit a pin payment confirmationrequest to the backend.
// A successful payment will result in the content being injected into temporal
func (api *API) submitPinPaymentConfirmation(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	paymentNumber, exists := c.GetPostForm("payment_number")
	if !exists {
		FailNoExistPostForm(c, "payment_number")
		return
	}
	txHash, exists := c.GetPostForm("tx_hash")
	if !exists {
		FailNoExistPostForm(c, "tx_hash")
		return
	}
	ppm := models.NewPaymentManager(api.DBM.DB)
	um := models.NewUserManager(api.DBM.DB)
	ethAddress, err := um.FindEthAddressByUserName(username)
	if err != nil {
		api.LogError(err, EthAddressSearchError)
		FailOnError(c, err)
		return
	}
	pp, err := ppm.FindPaymentByNumberAndAddress(paymentNumber, ethAddress)
	if err != nil {
		api.LogError(err, PaymentSearchError)
		FailOnError(c, err)
		return
	}
	mqURL := api.TConfig.RabbitMQ.URL

	ppc := queue.PinPaymentConfirmation{
		TxHash:        txHash,
		EthAddress:    ethAddress,
		PaymentNumber: paymentNumber,
		ContentHash:   pp.ObjectName,
	}
	qm, err := queue.Initialize(queue.PinPaymentConfirmationQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)
		FailOnError(c, err)
		return
	}
	fmt.Println("publishing message")
	if err = qm.PublishMessage(ppc); err != nil {
		api.LogError(err, QueuePublishError)
		FailOnError(c, err)
		return
	}
	api.Logger.WithFields(log.Fields{
		"service":        "api",
		"user":           username,
		"payment_number": paymentNumber,
	}).Info("pin payment confirmation being processed")

	Respond(c, http.StatusOK, gin.H{"response": pp})
}

// SubmitPaymentToContract is a highly "insecure" way of paying for TEMPORAL and essentially involves sending us a private key
func (api *API) submitPaymentToContract(c *gin.Context) {
	msg := fmt.Sprintf("this route requires you giving us your private key and the password to descrypt. Please provide a postform accept_warning set to yes otherwise this route will not work. Although we will not store your private key this is an extremely unsafe method as it means your private key can become compromised during transit or if someone where to gain control of our servers, and covertly save your key during usage. RTrade provides no insurance or protections against compromised accounts utilizing this route as it is intended for ADMIN USE ONLY or LAST RESORT USE ONLY. By using this route you full on agree that you void RTrade of any responsibilities, or fault that may occur as a resutl of your private key being compromised by using this route. DO NOT use this route if you do not agree with this")
	acceptWarn, exists := c.GetPostForm("accept_warning")
	if !exists {
		FailNoExistPostForm(c, "accept_warning")
		return
	}
	acceptWarn = strings.ToUpper(acceptWarn)
	switch acceptWarn {
	case "YES":
		break
	default:
		FailOnError(c, errors.New(msg))
		return

	}
	contentHash := c.Param("hash")
	username := GetAuthenticatedUserFromContext(c)
	holdTime, exists := c.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}
	keyFile, err := c.FormFile("key_file")
	if err != nil {
		FailOnError(c, err)
		return
	}
	keyFileHandler, err := keyFile.Open()
	if err != nil {
		api.LogError(err, FileOpenError)
		FailOnError(c, err)
		return
	}
	ethPass, exists := c.GetPostForm("eth_pass")
	if !exists {
		FailNoExistPostForm(c, "eth_pass")
		return
	}
	method, exists := c.GetPostForm("payment_method")
	if !exists {
		FailNoExistPostForm(c, "payment_method")
		return
	}
	methodUint, err := strconv.ParseUint(method, 10, 8)
	if err != nil {
		FailOnError(c, err)
		return
	}
	if methodUint > 1 {
		FailOnError(c, errors.New("payment_method must be 1 or 0"))
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}

	manager, err := rtfs.Initialize("", "")
	if err != nil {
		api.LogError(err, IPFSConnectionError)
		FailOnError(c, err)
		return
	}
	totalCost, err := utils.CalculatePinCost(contentHash, holdTimeInt, manager.Shell)
	if err != nil {
		api.LogError(err, PinCostCalculationError)
		FailOnError(c, err)
		return
	}
	costBig := utils.FloatToBigInt(totalCost)

	mqURL := api.TConfig.RabbitMQ.URL

	ppm := models.NewPaymentManager(api.DBM.DB)
	um := models.NewUserManager(api.DBM.DB)
	ethAddress, err := um.FindEthAddressByUserName(username)
	if err != nil {
		api.LogError(err, EthAddressSearchError)
		FailOnError(c, err)
		return
	}
	var number *big.Int
	num, err := ppm.RetrieveLatestPaymentNumberForUser(username)
	if err != nil {
		api.LogError(err, PaymentSearchError)
		FailOnError(c, err)
		return
	}
	if num.Cmp(big.NewInt(0)) == 1 {
		number = new(big.Int).Add(num, big.NewInt(1))
	} else {
		number = num
	}
	addressTyped := common.HexToAddress(ethAddress)
	ps, err := signer.GeneratePaymentSigner(
		api.TConfig.Ethereum.Account.KeyFile,
		api.TConfig.Ethereum.Account.KeyPass)
	if err != nil {
		api.LogError(err, PaymentSignerGenerationError)
		FailOnError(c, err)
		return
	}
	sm, err := ps.GenerateSignedPaymentMessagePrefixed(addressTyped, uint8(methodUint), number, costBig)
	if err != nil {
		api.LogError(err, PaymentMessageSignError)
		FailOnError(c, err)
		return
	}
	jsonKeyBytes, err := ioutil.ReadAll(keyFileHandler)
	if err != nil {
		// so in frequently used not logging
		FailOnError(c, err)
		return
	}
	pk, err := keystore.DecryptKey(jsonKeyBytes, ethPass)
	if err != nil {
		// so in frequently used not logging
		FailOnError(c, err)
		return
	}
	marshaledKey, err := pk.MarshalJSON()
	if err != nil {
		// so in frequently used not logging
		FailOnError(c, err)
		return
	}

	pps := queue.PinPaymentSubmission{
		PrivateKey:   marshaledKey,
		Method:       uint8(methodUint),
		Number:       number.String(),
		ChargeAmount: costBig.String(),
		ContentHash:  contentHash,
		H:            sm.H,
		V:            sm.V,
		R:            sm.R,
		S:            sm.S,
		Prefixed:     true,
		Hash:         sm.Hash,
		Sig:          sm.Sig,
	}

	if _, err = ppm.NewPayment(uint8(methodUint), number, costBig, ethAddress, contentHash, username, "pin", "public", holdTimeInt); err != nil {
		api.LogError(err, PaymentCreationError)
		FailOnError(c, err)
		return
	}
	qm, err := queue.Initialize(queue.PinPaymentSubmissionQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)
		FailOnError(c, err)
		return
	}

	if err = qm.PublishMessage(pps); err != nil {
		api.LogError(err, QueuePublishError)
		FailOnError(c, err)
		return
	}
	api.Logger.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("payment submitted to contract, user clearly ignored the warnings")

	Respond(c, http.StatusOK, gin.H{"response": pps})
}
