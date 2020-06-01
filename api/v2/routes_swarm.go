package v2

import (
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/RTradeLtd/database/v2/models"
	"github.com/gin-gonic/gin"
)

// SwarmUpload is used to upload data to ethereum swarm
func (api *API) SwarmUpload(c *gin.Context) {
	if len(api.swarmEndpoints) == 0 {
		Fail(c, errors.New("must have at least one swarm client"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	holdTime := c.PostForm("hold_time")
	holdTimeInMonthsInt, err := api.validateHoldTime(username, holdTime)
	if err != nil {
		Fail(c, err)
		return
	}
	fileHandler, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}
	fileSize := fileHandler.Size
	// ensure file size is within acceptable parameters
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		Fail(c, err)
		return
	}
	// calculate the cost of the file
	cost, err := utils.CalculateFileCost(username, holdTimeInMonthsInt, fileSize, api.usage)
	if err != nil {
		api.LogError(c, err, eh.CostCalculationError)(http.StatusBadRequest)
		return
	}
	// validate they have enough credits to pay for the upload
	if err = api.validateUserCredits(username, cost); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	// update their data usage
	if err := api.usage.UpdateDataUsage(username, uint64(fileSize)); err != nil {
		api.LogError(c, err, eh.CantUploadError)(http.StatusBadRequest)
		api.refundUserCredits(username, "file", cost)
		return
	}
	// now begin with the actual file uploading
	isTar := c.PostForm("is_tar")
	// open file handler
	openFile, err := fileHandler.Open()
	if err != nil {
		Fail(c, err)
		api.refundUserCredits(username, "file", cost)
		api.usage.ReduceDataUsage(username, uint64(fileSize))
		return
	}
	// read file data
	fileBytes, err := ioutil.ReadAll(openFile)
	if err != nil {
		Fail(c, err)
		api.refundUserCredits(username, "file", cost)
		api.usage.ReduceDataUsage(username, uint64(fileSize))
		return
	}
	// upload to both of our swarm nodes
	swarmHash, err := api.swarmUpload(fileBytes, isTar == "true")
	if err != nil {
		api.LogError(c, err, err.Error())
		api.refundUserCredits(username, "file", cost)
		api.usage.ReduceDataUsage(username, uint64(fileSize))
		return
	}
	// update uploads
	if _, err := api.upm.NewUpload(swarmHash, "swarm-file", models.UploadOptions{
		Username:         username,
		NetworkName:      "etherswarm",
		HoldTimeInMonths: holdTimeInMonthsInt,
		Size:             fileSize,
	}); err != nil {
		Fail(c, err)
		api.refundUserCredits(username, "file", cost)
		api.usage.ReduceDataUsage(username, uint64(fileSize))
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": swarmHash})
}
