package v2

import (
	"errors"
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
)

// getUserFromToken is used to get the username of the associated token
func (api *API) getUserFromToken(c *gin.Context) {
	// get username from jwt
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// return username
	Respond(c, http.StatusOK, gin.H{"response": username})
}

// verifyEmailAddress is used to verify a users email address
func (api *API) verifyEmailAddress(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract the token used to verify email
	forms := api.extractPostForms(c, "token")
	if len(forms) == 0 {
		return
	}
	// attempt email validation
	if _, err := api.um.ValidateEmailVerificationToken(username, forms["token"]); err != nil {
		api.LogError(c, err, eh.EmailVerificationError)(http.StatusBadRequest)
		return
	}
	// log and return
	Respond(c, http.StatusOK, gin.H{"response": "email verified"})
}

// getEmailVerificationToken is used to generate a token which can be used to validate an email address
func (api *API) getEmailVerificationToken(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// generate a random token to validate email
	user, err := api.um.GenerateEmailVerificationToken(username)
	if err != nil {
		api.LogError(c, err, eh.EmailTokenGenerationError)(http.StatusBadRequest)
		return
	}
	// build email message
	es := queue.EmailSend{
		Subject:     "TEMPORAL Email Verification",
		Content:     fmt.Sprintf("Please submit the following email verification token: %s\n", user.EmailVerificationToken),
		ContentType: "text/html",
		UserNames:   []string{user.UserName},
		Emails:      []string{user.EmailAddress},
	}
	// send email message to queue for processing
	if err = api.queues.email.PublishMessage(es); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// log and return
	Respond(c, http.StatusOK, gin.H{"response": "Email verification token sent to email. Please check and follow instructions"})
}

// ChangeAccountPassword is used to change a users password
func (api *API) changeAccountPassword(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract post forms
	forms := api.extractPostForms(c, "old_password", "new_password")
	if len(forms) == 0 {
		return
	}
	// parse html encoded strings
	forms["old_password"] = html.UnescapeString(forms["old_password"])
	forms["new_password"] = html.UnescapeString(forms["new_password"])
	api.l.With("user", username).Info("password change requested")
	// change password
	if ok, err := api.um.ChangePassword(username, forms["old_password"], forms["new_password"]); err != nil {
		api.LogError(c, err, eh.PasswordChangeError)(http.StatusBadRequest)
		return
	} else if !ok {
		err = fmt.Errorf("password changed failed for user %s to due an unspecified error", username)
		api.LogError(c, err, eh.PasswordChangeError)(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.Infow("password changed",
		"user", username)
	Respond(c, http.StatusOK, gin.H{"response": "password changed"})
}

// RegisterUserAccount is used to sign up with temporal
func (api *API) registerUserAccount(c *gin.Context) {
	// extract post forms
	forms := api.extractPostForms(c, "username", "password", "email_address")
	if len(forms) == 0 {
		return
	}
	// parse html encoded strings
	forms["password"] = html.UnescapeString(forms["password"])
	api.l.Info("user account registration detected")
	// create user model
	userModel, err := api.um.NewUserAccount(forms["username"], forms["password"], forms["email_address"])
	if err != nil {
		switch err.Error() {
		case eh.DuplicateEmailError:
			api.LogError(c, err, eh.DuplicateEmailError)(http.StatusBadRequest)
			return
		case eh.DuplicateUserNameError:
			api.LogError(c, err, eh.DuplicateUserNameError)(http.StatusBadRequest)
			return
		default:
			api.LogError(c, err, eh.UserAccountCreationError)(http.StatusBadRequest)
			return
		}
	}
	// generate a random token to validate email
	user, err := api.um.GenerateEmailVerificationToken(forms["username"])
	if err != nil {
		api.LogError(c, err, eh.EmailTokenGenerationError)(http.StatusBadRequest)
		return
	}
	// build email message
	es := queue.EmailSend{
		Subject:     "TEMPORAL Email Verification",
		Content:     fmt.Sprintf("Please submit the following email verification token: %s\n", user.EmailVerificationToken),
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
	api.l.With("user", forms["username"]).Info("user account registered")
	// remove hashed password from output
	userModel.HashedPassword = "scrubbed"
	// return
	Respond(c, http.StatusOK, gin.H{"response": userModel})
}

// CreateIPFSKey is used to create an IPFS key
func (api *API) createIPFSKey(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// extract forms
	forms := api.extractPostForms(c, "key_type", "key_bits", "key_name")
	if len(forms) == 0 {
		return
	}
	// validate key type
	switch forms["key_type"] {
	case "rsa", "ed25519":
		break
	default:
		// user error, do not log
		err := fmt.Errorf("%s is invalid key type must be rsa, or ed25519", forms["key_type"])
		Fail(c, err, http.StatusBadRequest)
		return
	}
	// search for user in database
	user, err := api.um.FindByUserName(username)
	if err != nil {
		api.LogError(c, err, eh.UserSearchError)(http.StatusNotFound)
		return
	}
	var cost float64
	// if they haven't made a key before, the first one is free
	if len(user.IPFSKeyIDs) == 0 {
		cost = 0
		err = nil
	} else {
		cost, err = utils.CalculateAPICallCost(forms["key_type"], false)
		if err != nil {
			api.LogError(c, err, eh.CallCostCalculationError)(http.StatusBadRequest)
			return
		}
	}
	// validate they have enough credits to pay for key creation
	if err := api.validateUserCredits(username, cost); err != nil && cost > 0 {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	// get a list of users current keys
	keys, err := api.um.GetKeysForUser(username)
	if err != nil {
		api.LogError(c, err, eh.KeySearchError)(http.StatusNotFound)
		api.refundUserCredits(username, "key", cost)
		return
	}
	// format key name
	// we prepend with the username to prevent key name collission
	keyName := fmt.Sprintf("%s-%s", username, forms["key_name"])
	// parse through existing key names, and ensure one doesnt' already exist
	for _, v := range keys["key_names"] {
		if v == keyName {
			err = fmt.Errorf("key with name already exists")
			api.LogError(c, err, eh.DuplicateKeyCreationError)(http.StatusConflict)
			api.refundUserCredits(username, "key", cost)
			return
		}
	}
	// parse key bits
	bitsInt, err := strconv.Atoi(forms["key_bits"])
	if err != nil {
		Fail(c, err)
		return
	}
	// create key creation message
	key := queue.IPFSKeyCreation{
		UserName:    username,
		Name:        keyName,
		Type:        forms["key_type"],
		Size:        bitsInt,
		NetworkName: "public",
		CreditCost:  cost,
	}
	// send message for processing
	if err = api.queues.key.PublishMessageWithExchange(key, queue.IpfsKeyExchange); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		api.refundUserCredits(username, "key", cost)
		return
	}
	// log and return
	api.l.With("user", username).Info("key creation request sent to backend")
	Respond(c, http.StatusOK, gin.H{"response": "key creation sent to backend"})
}

// GetIPFSKeyNamesForAuthUser is used to get the keys a user has setup
func (api *API) getIPFSKeyNamesForAuthUser(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// search for keys created by user
	keys, err := api.um.GetKeysForUser(username)
	if err != nil {
		api.LogError(c, err, eh.KeySearchError)(http.StatusBadRequest)
		return
	}
	// if the user has no keys, fail with an error
	if len(keys["key_names"]) == 0 || len(keys["key_ids"]) == 0 {
		Fail(c, errors.New(eh.NoKeyError), http.StatusNotFound)
		return
	}
	// log and return
	api.l.Infow("key name list requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": gin.H{"key_names": keys["key_names"], "key_ids": keys["key_ids"]}})
}

// GetCredits is used to get a users available credits
func (api *API) getCredits(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// get credits tied with the user account
	credits, err := api.um.GetCreditsForUser(username)
	if err != nil {
		api.LogError(c, err, eh.CreditCheckError)(http.StatusBadRequest)
		return
	}
	// log and return
	api.l.Infow("credit check requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": credits})
}

// ForgotEmail is used to retrieve an email if the user forgets it
func (api *API) forgotEmail(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	// search for the user
	user, err := api.um.FindByUserName(username)
	if err != nil {
		api.LogError(c, err, eh.UserSearchError)(http.StatusBadRequest)
		return
	}
	// return email address associated with the user account
	Respond(c, http.StatusOK, gin.H{"response": user.EmailAddress})
}

// ForgotUserName is used to send a username reminder to the email associated with the account
func (api *API) forgotUserName(c *gin.Context) {
	forms := api.extractPostForms(c, "email_address")
	if len(forms) == 0 {
		return
	}
	// find email address associated with the user account
	user, err := api.um.FindByEmail(forms["email_address"])
	if err != nil {
		Fail(c, errors.New(eh.UserSearchError), http.StatusBadRequest)
		return
	}
	// account email must be enabled in order to engage in account recovery processes
	if !user.EmailEnabled {
		Fail(c, errors.New("account does not have email enabled, unfortunately for security reasons we can't assist in recovery"))
		return
	}
	// construct email message
	es := queue.EmailSend{
		Subject:     "TEMPORAL User Name Reminder",
		Content:     fmt.Sprintf("your username is %s", user.UserName),
		ContentType: "text/html",
		UserNames:   []string{user.UserName},
		Emails:      []string{user.EmailAddress},
	}
	// send message for processing
	if err = api.queues.email.PublishMessage(es); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": "username reminder sent account email"})
}

// ResetPassword is used to reset the password associated with a user account
func (api *API) resetPassword(c *gin.Context) {
	forms := api.extractPostForms(c, "email_address")
	if len(forms) == 0 {
		return
	}
	// find user account associated with the email
	user, err := api.um.FindByEmail(forms["email_address"])
	if err != nil {
		api.LogError(c, err, eh.UserSearchError)(http.StatusBadRequest)
		return
	}
	// account email must be enabled in order to engage in account reovery process
	if !user.EmailEnabled {
		Fail(c, errors.New("account does not have email enabled, unfortunately for security reasons we can't assist in recovery"))
		return
	}
	// reset password, generating a random one
	newPass, err := api.um.ResetPassword(user.UserName)
	if err != nil {
		api.LogError(c, err, eh.PasswordResetError)(http.StatusBadRequest)
		return
	}
	// create email message
	es := queue.EmailSend{
		Subject:     "TEMPORAL Password Reset",
		Content:     fmt.Sprintf("your password is %s", newPass),
		ContentType: "text/html",
		UserNames:   []string{user.UserName},
		Emails:      []string{user.EmailAddress},
	}
	// send message to queue system for processing
	if err = api.queues.email.PublishMessage(es); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	// return
	Respond(c, http.StatusOK, gin.H{"response": "password reset, please check your email for a new password"})
}
