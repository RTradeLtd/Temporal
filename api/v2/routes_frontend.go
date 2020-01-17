package v2

import (
	"net/http"
	"strconv"

	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/RTradeLtd/database/v2/models"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

// CalculatePinCost is used to calculate the cost of pinning something to temporal
func (api *API) calculatePinCost(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// hash to calculate pin cost for
	hash := c.Param("hash")
	if _, err := gocid.Decode(hash); err != nil {
		Fail(c, err)
		return
	}
	// months to store pin for
	holdTime := c.Param("hold_time")
	// parse hold time
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	// calculate pin cost
	totalCost, err := utils.CalculatePinCost(username, hash, holdTimeInt, api.ipfs, api.usage)
	if err != nil {
		api.LogError(c, err, eh.CostCalculationError)(http.StatusBadRequest)
		Fail(c, err)
		return
	}
	// log and return
	api.l.With("user", username).Info("pin cost calculation requested")
	Respond(c, http.StatusOK, gin.H{"response": totalCost})
}

// CalculateFileCost is used to calculate the cost of uploading a file to our system
func (api *API) calculateFileCost(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// retrieve file to upload
	file, err := c.FormFile("file")
	if err != nil {
		Fail(c, err)
		return
	}
	// validate the file size is within limits
	if err := api.FileSizeCheck(file.Size); err != nil {
		Fail(c, err)
		return
	}
	// how many months to store file for
	forms, missingField := api.extractPostForms(c, "hold_time")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// parse hold time
	holdTimeInt, err := strconv.ParseInt(forms["hold_time"], 10, 64)
	if err != nil {
		Fail(c, err)
		return
	}
	api.l.With("user", username).Info("file cost calculation requested")
	// calculate cost
	cost, err := utils.CalculateFileCost(username, holdTimeInt, file.Size, api.usage)
	if err != nil {
		api.LogError(c, err, eh.CostCalculationError)(http.StatusBadRequest)
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": cost})
}

// GetEncryptedUploadsForUser is used to get all the encrypted uploads a user has
func (api *API) getEncryptedUploadsForUser(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	if c.Query("paged") == "true" {
		api.pageIt(c, api.ue.DB.Where("user_name = ?", username), &[]models.EncryptedUpload{})
		return
	}
	// find all uploads by this user
	uploads, err := api.ue.FindUploadsByUser(username)
	if err != nil && err != gorm.ErrRecordNotFound {
		api.LogError(c, err, eh.UploadSearchError)(http.StatusBadRequest)
		return
	}
	// if they haven't made any encrypted uploads, return a friendly message
	if len(*uploads) == 0 {
		Respond(c, http.StatusOK, gin.H{"response": "no encrypted uploads found, try them out :D"})
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": uploads})
}
