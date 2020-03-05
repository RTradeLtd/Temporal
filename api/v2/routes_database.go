package v2

import (
	"errors"
	"html"
	"net/http"
	"strings"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/database/v2/models"
	"github.com/gin-gonic/gin"
)

// allows performing arbitrary searches against upload file names
func (api *API) searchUploadsForUser(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms, missingField := api.extractPostForms(c, "search_query")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// escape string to prevent html encoded characters causing issues
	forms["search_query"] = html.UnescapeString(forms["search_query"])
	// force lower-case to make matching more likely s
	lower := strings.ToLower(forms["search_query"])
	// just in case lets try and avoid any possible headaches
	if strings.Contains(lower, "drop") || strings.Contains(lower, "select") {
		if strings.Contains(lower, "table") ||
			strings.Contains(lower, "column") ||
			strings.Contains(lower, "row") {
			Fail(c, errors.New("possible sql injection attack, goodbye"), http.StatusBadRequest)
			return
		}
	}
	if c.Query("paged") == "true" {
		api.pageIt(
			c,
			api.upm.DB.Where(
				"user_name = ? AND file_name_lower_case LIKE ?",
				username, lower,
			),
			&[]models.Upload{},
		)
		return
	}
	uploads, err := api.upm.Search(username, lower)
	if err != nil {
		api.LogError(c, err, eh.UploadSearchError)(http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": uploads})
}

// GetUploadsForUser is used to retrieve all uploads for the authenticated user
func (api *API) getUploadsForUser(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	if c.Query("paged") == "true" {
		api.pageIt(c, api.upm.DB.Where("user_name = ?", username), &[]models.Upload{})
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
		), &[]models.Upload{})
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
