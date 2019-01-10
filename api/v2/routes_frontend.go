package v2

import (
	"net/http"
	"strconv"

	"github.com/RTradeLtd/gorm"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

// CalculatePinCost is used to calculate the cost of pinning something to temporal
func (api *API) calculatePinCost(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	holdTime := c.Param("holdtime")
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	privateNet := c.PostForm("private_network")
	var isPrivate bool
	switch privateNet {
	case "true":
		isPrivate = true
	default:
		isPrivate = false
	}
	totalCost, err := utils.CalculatePinCost(hash, holdTimeInt, api.ipfs, isPrivate)
	if err != nil {
		api.LogError(err, eh.PinCostCalculationError)
		Fail(c, err)
		return
	}
	api.l.With("user", username).Info("pin cost calculation requested")
	Respond(c, http.StatusOK, gin.H{"response": totalCost})
}

// CalculateFileCost is used to calculate the cost of uploading a file to our system
func (api *API) calculateFileCost(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}
	if err := api.FileSizeCheck(file.Size); err != nil {
		Fail(c, err)
		return
	}
	forms := api.extractPostForms(c, "hold_time")
	if len(forms) == 0 {
		return
	}
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
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
	api.l.With("user", username).Info("file cost calculation requested")
	cost := utils.CalculateFileCost(holdTimeInt, file.Size, isPrivate)
	Respond(c, http.StatusOK, gin.H{"response": cost})
}

// GetEncryptedUploadsForUser is used to get all the encrypted uploads a user has
func (api *API) getEncryptedUploadsForUser(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	uploads, err := api.ue.FindUploadsByUser(username)
	if err != nil && err != gorm.ErrRecordNotFound {
		api.LogError(err, eh.UploadSearchError)(c, http.StatusBadRequest)
		return
	}
	if len(*uploads) == 0 {
		Respond(c, http.StatusOK, gin.H{"response": "no encrypted uploads found, try them out :D"})
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": uploads})
}
