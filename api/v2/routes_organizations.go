package v2

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/database/v2/models"
	gpaginator "github.com/RTradeLtd/gpaginator"
	"github.com/gin-gonic/gin"
	"github.com/jszwec/csvutil"
	"html"
	"net/http"
	"strconv"
	"time"
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
	if _, ok := api.validateOrgOwner(c, forms["organization_name"], username); !ok {
		return
	}
	// parse html encoded strings
	forms["password"] = html.UnescapeString(forms["password"])
	// create the org user. this process is similar to regular
	// user registration, so we handle the errors in the same way
	if _, err := api.orgs.RegisterOrgUser(
		forms["organization_name"],
		forms["username"],
		forms["password"],
		forms["email_address"],
	); err != nil {
		switch err.Error() {
		case eh.DuplicateEmailError:
			api.LogError(
				c,
				err,
				eh.DuplicateEmailError,
				"email",
				forms["email_address"])(http.StatusBadRequest)
			return
		case eh.DuplicateUserNameError:
			api.LogError(
				c,
				err,
				eh.DuplicateUserNameError,
				"username",
				forms["username"])(http.StatusBadRequest)
			return
		default:
			api.LogError(
				c,
				err,
				eh.UserAccountCreationError)(http.StatusBadRequest)
			return
		}
	}
	// generate a random token to validate email
	user, err := api.um.GenerateEmailVerificationToken(forms["username"])
	if err != nil {
		api.LogError(c, err, eh.EmailTokenGenerationError)(http.StatusBadRequest)
		return
	}
	// generate a jwt used to trigger email validation
	token, err := api.generateEmailJWTToken(user.UserName, user.EmailVerificationToken)
	if err != nil {
		api.LogError(c, err, "failed to generate email verification jwt")
		return
	}
	var url string
	// format the url the user clicks to activate email
	if dev {
		url = fmt.Sprintf(
			"https://dev.api.temporal.cloud/v2/account/email/verify/%s/%s",
			user.UserName, token,
		)
	} else {
		url = fmt.Sprintf(
			"https://api.temporal.cloud/v2/account/email/verify/%s/%s",
			user.UserName, token,
		)

	}
	// format a link tag
	link := fmt.Sprintf("<a href=\"%s\">link</a>", url)
	emailSubject := fmt.Sprintf(
		"%s Temporal Email Verification", forms["organization_name"],
	)
	// build email message
	es := queue.EmailSend{
		Subject: emailSubject,
		Content: fmt.Sprintf(
			"please click this %s to activate temporal email functionality", link,
		),
		ContentType: "text/html",
		UserNames:   []string{user.UserName},
		Emails:      []string{user.EmailAddress},
	}
	// send email message to queue for processing
	if err = api.queues.email.PublishMessage(es); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// log
	api.l.With(
		"user", forms["username"],
		"organization", forms["organization_name"],
		"organization.owner", username,
	).Info("organization user account registered")
	// remove hashed password from output
	user.HashedPassword = "scrubbed"
	// remove the verification token from output
	user.EmailVerificationToken = "scrubbed"
	// format a custom response that includes the user model
	// and an additional status field
	var status string
	if dev {
		status = fmt.Sprintf(
			"by continuing to use this service you agree to be bound by the following api terms and service %s",
			devTermsAndServiceURL,
		)
	} else {
		status = fmt.Sprintf(
			"by continuing to use this service you agree to be bound by the following api terms and service %s",
			prodTermsAndServiceURL,
		)
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": struct {
		*models.User
		Status string
	}{
		user, status,
	},
	})
}

// getOrgUserUploads allows returning uploads for organization users
// optionally
func (api *API) getOrgUserUploads(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	forms, missingField := api.extractPostForms(c, "name")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	users, ok := c.GetPostFormArray("users")
	if !ok {
		FailWithMissingField(c, "users")
		return
	}
	// allows optional returning the response as a generated csv file
	asCSV := c.PostForm("as_csv") == "true"
	// validate user is owner
	if _, ok := api.validateOrgOwner(c, forms["name"], username); !ok {
		return
	}
	if asCSV {
		type r struct {
			Users map[string][]models.Upload `json:"users"`
		}
		resp := &r{Users: make(map[string][]models.Upload)}
		uplds, err := api.getUploads(forms["name"], users)
		if err != nil {
			api.LogError(c, err, "failed to get user uploads"+err.Error())
			return
		}
		resp.Users = uplds
		csvBytes, err := csvutil.Marshal(resp)
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
	if !ok {
		Fail(c, err, http.StatusBadRequest)
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		Fail(c, err, http.StatusBadRequest)
		return
	}
	type pagedUploadsUser struct {
		Error        error             `json:"error,omitempty"`
		PagedUploads *gpaginator.Paged `json:"paged_uploads,omitempty"`
	}
	type pagedUploads struct {
		Users map[string]pagedUploadsUser `json:"users"`
	}
	// return paged uploads for all request users
	pu := &pagedUploads{
		Users: make(map[string]pagedUploadsUser),
	}
	for _, user := range users {
		// validate that the user is part of the organization
		// however dont fail on an error, simply continue
		usr, err := api.um.FindByUserName(user)
		if err != nil {
			continue
		}
		if usr.Organization != forms["name"] {
			continue
		}
		var (
			uploads []models.Upload
			pgu     pagedUploadsUser
		)
		paged, err := gpaginator.Paging(
			&gpaginator.Param{
				DB:    api.upm.DB.Where("user_name = ?", user),
				Page:  pageInt,
				Limit: limitInt,
			},
			&uploads,
		)
		if err != nil {
			pgu.Error = err
		} else {
			pgu.PagedUploads = paged
		}
		pu.Users[user] = pgu
	}
	// return the response
	Respond(c, http.StatusOK, gin.H{"response": pu})
}

func (api *API) getUploads(orgName string, users []string) (map[string][]models.Upload, error) {
	resp := make(map[string][]models.Upload)
	for _, user := range users {
		uplds, err := api.orgs.GetUserUploads(orgName, user)
		if err != nil {
			return nil, err
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
