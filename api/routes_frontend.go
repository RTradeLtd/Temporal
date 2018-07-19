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
	"github.com/jinzhu/gorm"
	minio "github.com/minio/minio-go"
)

/*
Contains routes used for frontend operation
*/

// CalculatePinCost is used to calculate the cost of pinning something to temporal
func CalculatePinCost(c *gin.Context) {
	hash := c.Param("hash")
	holdTime := c.Param("holdtime")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
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
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"total_cost_usd": totalCost,
	})
}

func CalculateFileCost(c *gin.Context) {
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
	cost := utils.CalculateFileCost(holdTimeInt, file.Size)
	c.JSON(http.StatusOK, gin.H{
		"total_cost_usd": cost,
	})
}

func CreatePinPayment(c *gin.Context) {
	contentHash := c.Param("hash")
	ethAddress := GetAuthenticatedUserFromContext(c)
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
		FailOnError(c, err)
		return
	}
	totalCost, err := utils.CalculatePinCost(contentHash, holdTimeInt, manager.Shell)
	if err != nil {
		FailOnError(c, err)
		return
	}
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}
	keyMap, ok := c.MustGet("eth_account").(map[string]string)
	if !ok {
		FailedToLoadMiddleware(c, "eth account")
		return
	}
	ps, err := signer.GeneratePaymentSigner(keyMap["keyFile"], keyMap["keyPass"])
	if err != nil {
		FailOnError(c, err)
		return
	}
	ppm := models.NewPinPaymentManager(db)
	var num *big.Int
	num, err = ppm.RetrieveLatestPaymentNumber(ethAddress)
	if err != nil && err != gorm.ErrRecordNotFound {
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
	costBig := utils.FloatToBigInt(totalCost)
	// for testing purpose
	addressTyped := common.HexToAddress(ethAddress)

	sm, err := ps.GenerateSignedPaymentMessagePrefixed(addressTyped, uint8(methodUint), num, costBig)
	if err != nil {
		FailOnError(c, err)
		return
	}

	_, err = ppm.NewPayment(uint8(methodUint), sm.PaymentNumber, sm.ChargeAmount, ethAddress, contentHash)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"h":                    sm.H,
		"v":                    sm.V,
		"r":                    sm.R,
		"s":                    sm.S,
		"eth_address":          sm.Address,
		"charge_amount_in_wei": sm.ChargeAmount,
		"payment_method":       sm.PaymentMethod,
		"payment_number":       sm.PaymentNumber,
	})
}

func CreateFilePayment(c *gin.Context) {
	cC := c.Copy()

	holdTimeInMonths, exists := cC.GetPostForm("hold_time")
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

	credentials, ok := cC.MustGet("minio_credentials").(map[string]string)
	if !ok {
		FailedToLoadMiddleware(c, "minio credentials")
		return
	}
	secure, ok := cC.MustGet("minio_secure").(bool)
	if !ok {
		FailedToLoadMiddleware(c, "minio secure")
		return
	}
	endpoint, ok := cC.MustGet("minio_endpoint").(string)
	if !ok {
		FailedToLoadMiddleware(c, "minio endpoint")
		return
	}
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}
	keyMap, ok := c.MustGet("eth_account").(map[string]string)
	if !ok {
		FailedToLoadMiddleware(c, "eth account")
		return
	}
	ps, err := signer.GeneratePaymentSigner(keyMap["keyFile"], keyMap["keyPass"])
	if err != nil {
		FailOnError(c, err)
		return
	}
	miniManager, err := mini.NewMinioManager(endpoint, credentials["access_key"], credentials["secret_key"], secure)
	if err != nil {
		FailOnError(c, err)
		return
	}
	fileHandler, err := cC.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("opening file")
	openFile, err := fileHandler.Open()
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("file opened")
	ethAddress := GetAuthenticatedUserFromContext(cC)

	holdTimeInMonthsInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}
	cost := utils.CalculateFileCost(holdTimeInMonthsInt, fileHandler.Size)
	costBig := utils.FloatToBigInt(cost)
	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", ethAddress, randString)
	fmt.Println("storing file in minio")
	_, err = miniManager.PutObject(FilesUploadBucket, objectName, openFile, fileHandler.Size, minio.PutObjectOptions{})
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("file stored in minio")

	fpm := models.NewFilePaymentManager(db)
	var num *big.Int
	num, err = fpm.RetrieveLatestPaymentNumber(ethAddress)
	if err != nil {
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
		FailOnError(c, err)
		return
	}
	_, err = fpm.NewPayment(uint8(methodUint), sm.PaymentNumber, sm.ChargeAmount, ethAddress, FilesUploadBucket, objectName)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"h":                    sm.H,
		"v":                    sm.V,
		"r":                    sm.R,
		"s":                    sm.S,
		"eth_address":          sm.Address,
		"charge_amount_in_wei": sm.ChargeAmount,
		"payment_method":       sm.PaymentMethod,
		"payment_number":       sm.PaymentNumber,
	})
}

func SubmitPinPaymentConfirmation(c *gin.Context) {
	ethAddress := GetAuthenticatedUserFromContext(c)
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
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}
	ppm := models.NewPinPaymentManager(db)
	pp, err := ppm.FindPaymentByNumberAndAddress(paymentNumber, ethAddress)
	if err != nil {
		FailOnError(c, err)
		return
	}
	mqURL, ok := c.MustGet("mq_conn_url").(string)
	if !ok {
		FailedToLoadMiddleware(c, "rabbitmq")
		return
	}
	ppc := queue.PinPaymentConfirmation{
		TxHash:        txHash,
		EthAddress:    ethAddress,
		PaymentNumber: paymentNumber,
		ContentHash:   pp.ContentHash,
	}
	qm, err := queue.Initialize(queue.PinPaymentConfirmationQueue, mqURL)
	if err != nil {
		FailOnError(c, err)
		return
	}
	fmt.Println("publishing message")
	err = qm.PublishMessage(ppc)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"payment": pp})
}

func SubmitPaymentToContract(c *gin.Context) {
	msg := fmt.Sprintf("this route requires you giving us your private key and the password to descrypt. Please provide a postform accept_warning set to yes otherwise this route will not work. Although we will not store your private key this is an extremely unsafe method as it means your private key can become compromised during transit or if someone where to gain control of our servers, and covertly save your key during usage. RTrade provides no insurance or protections against compromised accounts utilizing this route as it is intended for ADMIN USE ONLY or LAST RESORT USE ONLY. By using this route you full on agree that you void RTrade of any responsibilities, or fault that may occur as a resutl of your private key being compromised by using this route. DO NOT use this route if you do not agree with this")
	acceptWarn, exists := c.GetPostForm("accept_warning")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": msg,
		})
		return
	}
	acceptWarn = strings.ToUpper(acceptWarn)
	switch acceptWarn {
	case "YES":
		break
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "please set accept_warning to yes in order to continue",
		})
		return
	}
	contentHash := c.Param("hash")
	ethAddress := GetAuthenticatedUserFromContext(c)
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
		FailOnError(c, err)
		return
	}
	totalCost, err := utils.CalculatePinCost(contentHash, holdTimeInt, manager.Shell)
	if err != nil {
		FailOnError(c, err)
		return
	}
	costBig := utils.FloatToBigInt(totalCost)
	db, ok := c.MustGet("db").(*gorm.DB)
	if !ok {
		FailedToLoadDatabase(c)
		return
	}
	keyMap, ok := c.MustGet("eth_account").(map[string]string)
	if !ok {
		FailedToLoadMiddleware(c, "eth account")
		return
	}
	mqURL, ok := c.MustGet("mq_conn_url").(string)
	if !ok {
		FailedToLoadMiddleware(c, "rabbitmq")
		return
	}
	ppm := models.NewPinPaymentManager(db)
	var number *big.Int
	num, err := ppm.RetrieveLatestPaymentNumber(ethAddress)
	if err != nil {
		FailOnError(c, err)
		return
	}
	if num.Cmp(big.NewInt(0)) == 1 {
		number = new(big.Int).Add(num, big.NewInt(1))
	} else {
		number = num
	}
	addressTyped := common.HexToAddress(ethAddress)
	ps, err := signer.GeneratePaymentSigner(keyMap["keyFile"], keyMap["keyPass"])
	if err != nil {
		FailOnError(c, err)
		return
	}
	sm, err := ps.GenerateSignedPaymentMessagePrefixed(addressTyped, uint8(methodUint), number, costBig)
	if err != nil {
		FailOnError(c, err)
		return
	}
	jsonKeyBytes, err := ioutil.ReadAll(keyFileHandler)
	if err != nil {
		FailOnError(c, err)
		return
	}
	pk, err := keystore.DecryptKey(jsonKeyBytes, ethPass)
	if err != nil {
		FailOnError(c, err)
		return
	}
	marshaledKey, err := pk.MarshalJSON()
	if err != nil {
		FailOnError(c, err)
		return
	}
	smm := make(map[string]interface{})
	smm["h"] = sm.H
	smm["v"] = sm.V
	smm["r"] = sm.R
	smm["s"] = sm.S

	pps := queue.PinPaymentSubmission{
		PrivateKey:    marshaledKey,
		Method:        uint8(methodUint),
		Number:        number.String(),
		ChargeAmount:  costBig.String(),
		ContentHash:   contentHash,
		SignedMessage: smm,
	}

	qm, err := queue.Initialize(queue.PinPaymentSubmissionQueue, mqURL)
	if err != nil {
		FailOnError(c, err)
		return
	}

	err = qm.PublishMessage(pps)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": pps,
	})
}
