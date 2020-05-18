package v2

import (
	"io/ioutil"
	"net/http"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/database/v2/models"
	"github.com/gin-gonic/gin"
)

// SwarmUpload is used to upload data to ethereum swarm
func (api *API) SwarmUpload(c *gin.Context) {
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
	size := fileHandler.Size
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		Fail(c, err)
		return
	}
	isTar := c.PostForm("is_tar")
	openFile, err := fileHandler.Open()
	if err != nil {
		Fail(c, err)
		return
	}
	fileBytes, err := ioutil.ReadAll(openFile)
	if err != nil {
		Fail(c, err)
		return
	}
	// TODO(bonedaddy): bill user
	swarmHash, err := api.dualSwarmUpload(fileBytes, isTar == "true")
	if err != nil {
		api.LogError(c, err, err.Error())
		return
	}

	if _, err := api.upm.NewUpload(swarmHash, "swarm-file", models.UploadOptions{
		Username:         username,
		NetworkName:      "etherswarm",
		HoldTimeInMonths: holdTimeInMonthsInt,
		Size:             size,
	}); err != nil {
		Fail(c, err)
		// TODO(bonedaddy): refund credits if this fails
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": swarmHash})
}
