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

	api.l.WithFields(log.Fields{
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
		FailOnError(c, err)
		return
	}
	if err := api.FileSizeCheck(file.Size); err != nil {
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

	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("file cost calculation requested")

	cost := utils.CalculateFileCost(holdTimeInt, file.Size)
	Respond(c, http.StatusOK, gin.H{"response": cost})
}
