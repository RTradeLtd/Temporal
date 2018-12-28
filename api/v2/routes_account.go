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
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": username})
}

// verifyEmailAddress is used to verify a users email address
func (api *API) verifyEmailAddress(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "token")
	if len(forms) == 0 {
		return
	}
	if _, err := api.um.ValidateEmailVerificationToken(username, forms["token"]); err != nil {
		api.LogError(err, eh.EmailVerificationError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "email verified"})
}

// getEmailVerificationToken is used to generate a token which can be used to validate an email address
func (api *API) getEmailVerificationToken(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	user, err := api.um.GenerateEmailVerificationToken(username)
	if err != nil {
		api.LogError(err, eh.EmailTokenGenerationError)(c, http.StatusBadRequest)
		return
	}
	es := queue.EmailSend{
		Subject:     "TEMPORAL Email Verification",
		Content:     fmt.Sprintf("Please submit the following email verification token: %s\n", user.EmailVerificationToken),
		ContentType: "text/html",
		UserNames:   []string{user.UserName},
		Emails:      []string{user.EmailAddress},
	}
	if err = api.queues.email.PublishMessage(es); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "Email verification token sent to email. Please check and follow instructions"})
}

// ChangeAccountPassword is used to change a users password
func (api *API) changeAccountPassword(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "old_password", "new_password")
	if len(forms) == 0 {
		return
	}
	forms["old_password"] = html.UnescapeString(forms["old_password"])
	forms["new_password"] = html.UnescapeString(forms["new_password"])
	api.l.With("user", username).Info("password change requested")
	if ok, err := api.um.ChangePassword(username, forms["old_password"], forms["new_password"]); err != nil {
		api.LogError(err, eh.PasswordChangeError)(c, http.StatusBadRequest)
		return
	} else if !ok {
		err = fmt.Errorf("password changed failed for user %s to due an unspecified error", username)
		api.LogError(err, eh.PasswordChangeError)(c)
		return
	}
	api.l.Infow("password changed",
		"user", username)
	Respond(c, http.StatusOK, gin.H{"response": "password changed"})
}

// RegisterUserAccount is used to sign up with temporal
func (api *API) registerUserAccount(c *gin.Context) {
	forms := api.extractPostForms(c, "username", "password", "email_address")
	if len(forms) == 0 {
		return
	}
	forms["password"] = html.UnescapeString(forms["password"])
	api.l.Info("user account registration detected")
	userModel, err := api.um.NewUserAccount(forms["username"], forms["password"], forms["email_address"])
	if err != nil {
		switch err.Error() {
		case eh.DuplicateEmailError:
			api.LogError(err, eh.DuplicateEmailError)(c, http.StatusBadRequest)
			return
		case eh.DuplicateUserNameError:
			api.LogError(err, eh.DuplicateUserNameError)(c, http.StatusBadRequest)
			return
		default:
			api.LogError(err, eh.UserAccountCreationError)(c, http.StatusBadRequest)
			return
		}
	}
	api.l.With("user", forms["username"]).Info("user account registered")
	userModel.HashedPassword = "scrubbed"
	Respond(c, http.StatusOK, gin.H{"response": userModel})
}

// CreateIPFSKey is used to create an IPFS key
func (api *API) createIPFSKey(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	forms := api.extractPostForms(c, "key_type", "key_bits", "key_name")
	if len(forms) == 0 {
		return
	}
	switch forms["key_type"] {
	case "rsa", "ed25519":
		break
	default:
		// user error, do not log
		err := fmt.Errorf("%s is invalid key type must be rsa, or ed25519", forms["key_type"])
		Fail(c, err, http.StatusBadRequest)
		return
	}
	user, err := api.um.FindByUserName(username)
	if err != nil {
		api.LogError(err, eh.UserSearchError)(c, http.StatusNotFound)
		return
	}
	var cost float64
	// if they haven't made a key before, the first one is free
	if len(user.IPFSKeyIDs) == 0 {
		cost = 0
		err = nil
	} else {
		cost, err = utils.CalculateAPICallCost(forms["key_type"], false)
	}
	if err != nil {
		api.LogError(err, eh.CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil && cost > 0 {
		api.LogError(err, eh.InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	keys, err := api.um.GetKeysForUser(username)
	if err != nil {
		api.LogError(err, eh.KeySearchError)(c, http.StatusNotFound)
		api.refundUserCredits(username, "key", cost)
		return
	}
	// prefix the key name for validation, and further processing
	keyName := fmt.Sprintf("%s-%s", username, forms["key_name"])
	for _, v := range keys["key_names"] {
		if v == keyName {
			err = fmt.Errorf("key with name already exists")
			api.LogError(err, eh.DuplicateKeyCreationError)(c, http.StatusConflict)
			api.refundUserCredits(username, "key", cost)
			return
		}
	}
	bitsInt, err := strconv.Atoi(forms["key_bits"])
	if err != nil {
		Fail(c, err)
		return
	}
	key := queue.IPFSKeyCreation{
		UserName:    username,
		Name:        keyName,
		Type:        forms["key_type"],
		Size:        bitsInt,
		NetworkName: "public",
		CreditCost:  cost,
	}
	if err = api.queues.key.PublishMessageWithExchange(key, queue.IpfsKeyExchange); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		api.refundUserCredits(username, "key", cost)
		return
	}
	api.l.With("user", username).Info("key creation request sent to backend")
	Respond(c, http.StatusOK, gin.H{"response": "key creation sent to backend"})
}

// GetIPFSKeyNamesForAuthUser is used to get the keys a user has setup
func (api *API) getIPFSKeyNamesForAuthUser(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	keys, err := api.um.GetKeysForUser(username)
	if err != nil {
		api.LogError(err, eh.KeySearchError)(c)
		return
	}
	// if the user has no keys, fail with an error
	if len(keys["key_names"]) == 0 || len(keys["key_ids"]) == 0 {
		Fail(c, errors.New(eh.NoKeyError), http.StatusNotFound)
		return
	}
	api.l.Infow("key name list requested", "user", username)

	Respond(c, http.StatusOK, gin.H{"response": gin.H{"key_names": keys["key_names"], "key_ids": keys["key_ids"]}})
}

// GetCredits is used to get a users available credits
func (api *API) getCredits(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	credits, err := api.um.GetCreditsForUser(username)
	if err != nil {
		api.LogError(err, eh.CreditCheckError)(c, http.StatusBadRequest)
		return
	}
	api.l.Infow("credit check requested", "user", username)
	Respond(c, http.StatusOK, gin.H{"response": credits})
}

// ForgotEmail is used to retrieve an email if the user forgets it
func (api *API) forgotEmail(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	user, err := api.um.FindByUserName(username)
	if err != nil {
		api.LogError(err, eh.UserSearchError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": user.EmailAddress})
}

// ForgotUserName is used to send a username reminder to the email associated with the account
func (api *API) forgotUserName(c *gin.Context) {
	forms := api.extractPostForms(c, "email_address")
	if len(forms) == 0 {
		return
	}
	user, err := api.um.FindByEmail(forms["email_address"])
	if err != nil {
		Fail(c, errors.New(eh.UserSearchError), http.StatusBadRequest)
		return
	}
	if !user.EmailEnabled {
		Fail(c, errors.New("account does not have email enabled, unfortunately for security reasons we can't assist in recovery"))
		return
	}
	es := queue.EmailSend{
		Subject:     "TEMPORAL User Name Reminder",
		Content:     fmt.Sprintf("your username is %s", user.UserName),
		ContentType: "text/html",
		UserNames:   []string{user.UserName},
		Emails:      []string{user.EmailAddress},
	}
	if err = api.queues.email.PublishMessage(es); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "username reminder sent account email"})
}

// ResetPassword is used to reset the password associated with a user account
func (api *API) resetPassword(c *gin.Context) {
	forms := api.extractPostForms(c, "email_address")
	if len(forms) == 0 {
		return
	}
	user, err := api.um.FindByEmail(forms["email_address"])
	if err != nil {
		api.LogError(err, eh.UserSearchError)(c, http.StatusBadRequest)
		return
	}
	if !user.EmailEnabled {
		Fail(c, errors.New("account does not have email enabled, unfortunately for security reasons we can't assist in recovery"))
		return
	}
	newPass, err := api.um.ResetPassword(user.UserName)
	if err != nil {
		api.LogError(err, eh.PasswordResetError)(c, http.StatusBadRequest)
		return
	}
	es := queue.EmailSend{
		Subject:     "TEMPORAL Password Reset",
		Content:     fmt.Sprintf("your password is %s", newPass),
		ContentType: "text/html",
		UserNames:   []string{user.UserName},
		Emails:      []string{user.EmailAddress},
	}
	if err = api.queues.email.PublishMessage(es); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "password reset, please check your email for a new password"})
}
