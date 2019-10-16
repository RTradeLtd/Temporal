package v2

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/gin-gonic/gin"
)

// creates a new organization
func (api *API) newOrganization(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get the organization name
	forms, missingField := api.extractPostForms(c, "name")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// create the organization
	if err := api.orgs.NewOrganization(
		forms["name"],
		username,
	); err != nil {
		// creation failed, send an error message
		api.LogError(
			c,
			err,
			"failed to create organization",
		)(http.StatusInternalServerError)
		return
	}
	api.l.Infow("organization created",
		"name", forms["name"], "owner", username)
	Respond(c, http.StatusOK, gin.H{"response": "organization created"})
}
