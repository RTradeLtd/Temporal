package api

import (
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
	log "github.com/sirupsen/logrus"
)

// CalculatePinCost is used to calculate the cost of pinning something to temporal
func (api *API) calculatePinCost(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	holdTime := c.Param("holdtime")
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		api.LogError(err, IPFSConnectionError)(c)
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	privateNetwork := c.PostForm("private_network")
	var isPrivate bool
	switch privateNetwork {
	case "true":
		isPrivate = true
	default:
		isPrivate = false
	}
	totalCost, err := utils.CalculatePinCost(hash, holdTimeInt, manager.Shell, isPrivate)
	if err != nil {
		api.LogError(err, PinCostCalculationError)
		Fail(c, err)
		return
	}

	api.l.WithFields(log.Fields{
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
		Fail(c, err)
		return
	}
	if err := api.FileSizeCheck(file.Size); err != nil {
		Fail(c, err)
		return
	}
	holdTime, exists := c.GetPostForm("hold_time")
	if !exists {
		FailWithMissingField(c, "hold_time")
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	privateNetwork := c.PostForm("private_network")
	var isPrivate bool
	switch privateNetwork {
	case "true":
		isPrivate = true
	default:
		isPrivate = false
	}
	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("file cost calculation requested")

	cost := utils.CalculateFileCost(holdTimeInt, file.Size, isPrivate)
	Respond(c, http.StatusOK, gin.H{"response": cost})
}
