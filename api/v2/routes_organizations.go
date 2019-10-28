package v2

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/database/v2/models"
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
	org, err := api.orgs.FindByName(forms["name"])
	if err != nil {
		api.LogError(
			c,
			err,
			"failed to find org",
		)(http.StatusInternalServerError)
		return
	}
	if org.AccountOwner != username {
		api.LogError(
			c,
			errors.New("user is not owner"),
			"you are not the organization owner",
		)(http.StatusForbidden)
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
	// validate user is owner
	if org, err := api.orgs.FindByName(forms["name"]); err != nil {
		api.LogError(
			c,
			err,
			"failed to find org",
		)(http.StatusInternalServerError)
		return
	} else if org.AccountOwner != username {
		api.LogError(
			c,
			errors.New("user is not owner"),
			"you are not the organization owner",
		)(http.StatusForbidden)
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
	// ensure user is org owner
	org, err := api.orgs.FindByName(forms["organization_name"])
	if err != nil {
		api.LogError(
			c,
			err,
			"failed to find organization",
		)(http.StatusInternalServerError)
		return
	}
	if org.AccountOwner != username {
		api.LogError(
			c,
			errors.New("user is not owner"),
			"you are not the organization owner",
		)(http.StatusForbidden)
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
