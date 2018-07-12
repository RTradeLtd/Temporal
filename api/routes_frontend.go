package api

import (
	"errors"
	"math/big"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/payment_server/signer"
	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
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

	sm, err := ps.GenerateSignedPaymentMessage(addressTyped, uint8(methodUint), number, costBig)
	if err != nil {
		FailOnError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"signed_message": sm,
	})
}
