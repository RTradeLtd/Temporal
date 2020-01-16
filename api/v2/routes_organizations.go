package v2

import (
	"bytes"
	"errors"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/database/v2/models"
	gpaginator "github.com/RTradeLtd/gpaginator"
	"github.com/gin-gonic/gin"
	"github.com/jszwec/csvutil"
)

// creates a new organization
func (api *API) newOrganization(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// first validate they are a partner
	if org, err := api.usage.FindByUserName(username); err != nil {
		api.LogError(c, err, eh.UserSearchError)(http.StatusBadRequest)
		return
	} else if org.Tier != models.Partner {
		api.LogError(
			c,
			errors.New("account tier is not partner"),
			"only partner accounts can create orgs",
		)(http.StatusBadRequest)
		return
	}
	// get the organization name
	forms, missingField := api.extractPostForms(c, "name")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// create the organization
	if _, err := api.orgs.NewOrganization(
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

// getOrganization returns the organization model
// can only be called by organization owner
func (api *API) getOrganization(c *gin.Context) {
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
	org, ok := api.validateOrgOwner(c, forms["name"], username)
	if !ok {
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": org})
}

func (api *API) getOrgBillingReport(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get the organization name
	forms, missingField := api.extractPostForms(c, "name", "number_of_days")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// validate number_of_days parameter
	numDays, err := strconv.Atoi(forms["number_of_days"])
	if err != nil {
		api.LogError(
			c,
			err,
			"number_of_days is not an int",
		)(http.StatusBadRequest)
		return
	}
	if _, ok := api.validateOrgOwner(c, forms["name"], username); !ok {
		return
	}
	// generate a billing report
	report, err := api.orgs.GenerateBillingReport(
		forms["name"],
		time.Now().AddDate(0, 0, -numDays),
		time.Now(),
	)
	if err != nil {
		api.LogError(
			c,
			err,
			"failed to generate billing report",
		)(http.StatusInternalServerError)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": report})
}

// registerOrgUser is used to register an organization user
// unlike regular user registration, we dont check catch all
// email addresses
func (api *API) registerOrgUser(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract post forms
	forms, missingField := api.extractPostForms(
		c,
		"username",
		"password",
		"email_address",
		"organization_name",
	)
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// prevent people from registering usernames that contain an `@` sign
	// this prevents griefing by prevent user sign-ins by using a username
	// that is based off an email address
	if strings.ContainsRune(forms["username"], '@') {
		Fail(c, errors.New("usernames cant contain @ sign"))
		return
	}
	if _, ok := api.validateOrgOwner(c, forms["organization_name"], username); !ok {
		return
	}
	// parse html encoded strings
	forms["password"] = html.UnescapeString(forms["password"])
	// create the org user. this process is similar to regular
	// user registration, so we handle the errors in the same way
	_, err = api.orgs.RegisterOrgUser(
		forms["organization_name"],
		forms["username"],
		forms["password"],
		forms["email_address"],
	)
	api.handleUserCreate(c, forms, err)
}

// getOrgUserUploads allows returning uploads for organization users
// optionally
func (api *API) getOrgUserUploads(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms, missingField := api.extractPostForms(c, "name", "user")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// allows optional returning the response as a generated csv file
	asCSV := c.PostForm("as_csv") == "true"
	// validate user is owner
	if _, ok := api.validateOrgOwner(c, forms["name"], username); !ok {
		return
	}
	if asCSV {
		uplds, err := api.getUploads(forms["name"], []string{forms["user"]})
		if err != nil {
			api.LogError(c, err, "failed to get user uploads"+err.Error())
			return
		}
		csvBytes, err := csvutil.Marshal(uplds[forms["user"]])
		if err != nil {
			api.LogError(c, err, "failed to generate csv file "+err.Error())
			return
		}
		c.DataFromReader(
			200,
			int64(len(csvBytes)),
			"application/octet-stream",
			bytes.NewReader(csvBytes),
			make(map[string]string),
		)
		return
	}
	page := c.PostForm("page")
	if page == "" {
		page = "1"
	}
	limit := c.PostForm("limit")
	if limit == "" {
		limit = "10"
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		Fail(c, err, http.StatusBadRequest)
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		Fail(c, err, http.StatusBadRequest)
		return
	}
	// validate that the user is part of the organization
	// however dont fail on an error, simply continue
	usr, err := api.um.FindByUserName(forms["user"])
	if err != nil {
		api.LogError(c, err, eh.UserSearchError)
		return
	}
	if usr.Organization != forms["name"] {
		Fail(c, errors.New("user is not part of organization"))
		return
	}
	var uploads []models.Upload
	paged, err := gpaginator.Paging(
		&gpaginator.Param{
			DB:    api.upm.DB.Where("user_name = ?", forms["user"]),
			Page:  pageInt,
			Limit: limitInt,
		},
		&uploads,
	)
	if err != nil {
		api.LogError(c, err, "failed to get paged user upload")
		return
	}
	// return the response
	Respond(c, http.StatusOK, gin.H{"response": paged})
}

func (api *API) getUploads(orgName string, users []string) (map[string][]models.Upload, error) {
	resp := make(map[string][]models.Upload)
	for _, user := range users {
		uplds, err := api.orgs.GetUserUploads(orgName, user)
		if err != nil {
			continue
		}
		resp[user] = uplds
	}
	return resp, nil
}

// returns true if user is owner
func (api *API) validateOrgOwner(c *gin.Context, organization, username string) (*models.Organization, bool) {
	org, err := api.orgs.FindByName(organization)
	if err != nil {
		api.LogError(
			c,
			err,
			"failed to find org",
		)(http.StatusInternalServerError)
		return nil, false
	}
	if org.AccountOwner != username {
		api.LogError(
			c,
			errors.New("user is not owner"),
			"you are not the organization owner",
		)(http.StatusForbidden)
		return nil, false
	}
	return org, true
}
