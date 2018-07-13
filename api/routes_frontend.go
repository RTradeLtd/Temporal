package api

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/signer"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
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
	costBig := utils.FloatToBigInt(totalCost)
	// for testing purpose
	number := big.NewInt(0)
	addressTyped := common.HexToAddress(ethAddress)

	sm, err := ps.GenerateSignedPaymentMessagePrefixed(addressTyped, uint8(methodUint), number, costBig)
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
	mqURL, ok := c.MustGet("mq_conn_url").(string)
	if !ok {
		FailedToLoadMiddleware(c, "rabbitmq")
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
	ifp := queue.IPFSFile{
		BucketName:       FilesUploadBucket,
		ObjectName:       objectName,
		EthAddress:       ethAddress,
		NetworkName:      "public",
		HoldTimeInMonths: holdTimeInMonths,
	}
	qm, err := queue.Initialize(queue.IpfsFileQueue, mqURL)
	if err != nil {
		FailOnError(c, err)
		return
	}

	err = qm.PublishMessage(ifp)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"eth_address": ethAddress,
		"hold_time":   holdTimeInMonthsInt,
	})
}
