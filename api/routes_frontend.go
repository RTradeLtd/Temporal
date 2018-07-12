package api

import (
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/utils"

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
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 640)
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
}
