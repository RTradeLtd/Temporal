package v2

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/database/v2/models"
	"github.com/gin-gonic/gin"
)

// GetUploadsForUser is used to retrieve all uploads for the authenticated user
func (api *API) getUploadsForUser(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	if c.Query("paged") == "true" {
		api.pageIt(c, api.upm.DB.Where("user_name = ?", username), []models.Upload{})
		return
	}
	// fetch all uploads by the specified user
	uploads, err := api.upm.GetUploadsForUser(username)
	if err != nil {
		api.LogError(c, err, eh.UploadSearchError)(http.StatusInternalServerError)
		return
	}
	// log and return
	api.l.Info("specific uploads from database requested")
	Respond(c, http.StatusOK, gin.H{"response": uploads})
}

// getUploadsByNetworkName is used to get uploads for a network by its name
func (api *API) getUploadsByNetworkName(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get network name to retrieve uploads from
	networkName := c.Param("networkName")
	// validate the user can access the network
	if err := CheckAccessForPrivateNetwork(username, networkName, api.dbm.DB); err != nil {
		api.LogError(c, err, eh.PrivateNetworkAccessError)(http.StatusBadRequest)
		return
	}
	if c.Query("paged") == "true" {
		api.pageIt(c, api.upm.DB.Where(
			"user_name = ? AND network_name = ?",
			username, networkName,
		), []models.Upload{})
		return
	}
	// find uploads for the network
	uploads, err := api.upm.FindUploadsByNetwork(networkName)
	if err != nil {
		api.LogError(c, err, eh.UploadSearchError)(http.StatusInternalServerError)
		return
	}
	// log and return
	api.l.Infow("uploads forprivate ifps network requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": uploads})
}
