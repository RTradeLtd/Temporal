package api

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// getUserFromToken is used to get the username of the associated token
func (api *API) getUserFromToken(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	Respond(c, http.StatusOK, gin.H{"response": username})
}

// selfRekt is an undocumented API call used to auto-ban users who may engaging in malicious activity
func (api *API) selfRekt(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	user, err := api.um.FindByUserName(username)
	if err != nil {
		api.LogError(err, eh.UserSearchError)(c, http.StatusBadRequest)
		return
	}
	user.AccountEnabled = false
	user.APIAccess = false
	user.AdminAccess = false
	user.EnterpriseEnabled = false
	if err = api.dbm.DB.Save(user).Error; err != nil {
		api.LogError(err, eh.UnableToSaveUserError)(c, http.StatusBadRequest)
		return
	}
	api.LogWithUser(username).Info("malicious activity detected")
	Respond(c, http.StatusOK, gin.H{"response": "4 hour ban ... you been messin around with our shit, aint you son?"})
}

// verifyEmailAddress is used to verify a users email address
func (api *API) verifyEmailAddress(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
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
	username := GetAuthenticatedUserFromContext(c)
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
	mqURL := api.cfg.RabbitMQ.URL
	qm, err := queue.Initialize(queue.EmailSendQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c, http.StatusBadRequest)
		return
	}
	if err = qm.PublishMessage(es); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "Email verification token sent to email. Please check and follow instructions"})
}

// registerAirDrop is used to register an airdrop
func (api *API) registerAirDrop(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	forms := api.extractPostForms(c, "airdrop_id", "eth_address")
	if len(forms) == 0 {
		return
	}
	if len(forms["eth_address"]) != 42 {
		Fail(c, errors.New("eth_address is invalid"))
		return
	}
	if _, err := api.dm.RegisterAirDrop(forms["airdrop_id"], forms["eth_address"], username); err != nil {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	mqURL := api.cfg.RabbitMQ.URL
	qm, err := queue.Initialize(queue.EmailSendQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c, http.StatusBadRequest)
		return
	}
	es := queue.EmailSend{
		Subject:     "airdrop registered",
		Content:     fmt.Sprintf("user %s with airdrop code %s to eth address %s is registered", username, forms["eth_address"], forms["aidrop_id"]),
		ContentType: "text/html",
		UserNames:   []string{"admin"},
		Emails:      []string{os.Getenv("TEMPORAL_ADMIN_EMAIL")},
	}
	if err = qm.PublishMessage(es); err != nil {
		api.LogError(err, eh.QueuePublishError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": "airdrop registered, good luck!"})
}

// ChangeAccountPassword is used to change a users password
func (api *API) changeAccountPassword(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	forms := api.extractPostForms(c, "old_assword", "new_password")
	if len(forms) == 0 {
		return
	}
	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("password change requested")
	suceeded, err := api.um.ChangePassword(username, forms["old_password"], forms["new_password"])
	if err != nil {
		api.LogError(err, eh.PasswordChangeError)(c)
		return
	}
	if !suceeded {
		err = fmt.Errorf("password changed failed for user %s to due an unspecified error", username)
		api.LogError(err, eh.PasswordChangeError)(c)
		return
	}
	api.l.WithFields(log.Fields{
		"service": api,
		"user":    username,
	}).Info("password changed")
	Respond(c, http.StatusOK, gin.H{"response": "password changed"})
}

// RegisterUserAccount is used to sign up with temporal
func (api *API) registerUserAccount(c *gin.Context) {
	forms := api.extractPostForms(c, "username", "password", "email_address")
	if len(forms) == 0 {
		return
	}
	api.l.WithFields(log.Fields{
		"service": "api",
	}).Info("user account registration detected")
	userModel, err := api.um.NewUserAccount(forms["username"], forms["password"], forms["email_address"], false)
	if err != nil && err.Error() == eh.DuplicateEmailError || err != nil && err.Error() == eh.DuplicateUserNameError {
		api.LogError(err, err.Error())(c, http.StatusBadRequest)
		return
	}
	if err != nil {
		api.LogError(err, eh.UserAccountCreationError)(c, http.StatusBadRequest)
		return
	}
	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    forms["username"],
	}).Info("user account registered")
	userModel.HashedPassword = "scrubbed"
	Respond(c, http.StatusOK, gin.H{"response": userModel})
}

// CreateIPFSKey is used to create an IPFS key
func (api *API) createIPFSKey(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	forms := api.extractPostForms(c, "key_type", "key_bits", "key_name")
	if len(forms) == 0 {
		return
	}
	switch forms["key_type"] {
	case "rsa":
		break
	case "ed25519":
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
		if forms["key_type"] == "rsa" {
			cost, err = utils.CalculateAPICallCost("rsa-key", false)
		} else {
			cost, err = utils.CalculateAPICallCost("ed-key", false)
		}
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
	keyNamePrefixed := fmt.Sprintf("%s-%s", username, forms["key_name"])
	for _, v := range keys["key_names"] {
		if v == keyNamePrefixed {
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
	// if key type is RSA, and size is too small or too large, default to an appropriately size minimum
	if forms["key_type"] == "rsa" {
		if bitsInt > 4096 || bitsInt < 2048 {
			bitsInt = 2048
		}
	}
	key := queue.IPFSKeyCreation{
		UserName:    username,
		Name:        forms["key_name"],
		Type:        forms["key_type"],
		Size:        bitsInt,
		NetworkName: "public",
		CreditCost:  cost,
	}
	mqConnectionURL := api.cfg.RabbitMQ.URL
	qm, err := queue.Initialize(queue.IpfsKeyCreationQueue, mqConnectionURL, true, false)
	if err != nil {
		api.LogError(err, eh.QueueInitializationError)(c)
		api.refundUserCredits(username, "key", cost)
		return
	}
	if err = qm.PublishMessageWithExchange(key, queue.IpfsKeyExchange); err != nil {
		api.LogError(err, eh.QueuePublishError)(c)
		api.refundUserCredits(username, "key", cost)
		return
	}
	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("key creation request sent to backend")
	Respond(c, http.StatusOK, gin.H{"response": "key creation sent to backend"})
}

// GetIPFSKeyNamesForAuthUser is used to get the keys a user has setup
func (api *API) getIPFSKeyNamesForAuthUser(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)

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
	api.LogWithUser(username).Info("key name list requested")

	Respond(c, http.StatusOK, gin.H{"response": gin.H{"key_names": keys["key_names"], "key_ids": keys["key_ids"]}})
}

// GetCredits is used to get a users available credits
func (api *API) getCredits(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	credits, err := api.um.GetCreditsForUser(username)
	if err != nil {
		api.LogError(err, eh.CreditCheckError)(c, http.StatusBadRequest)
		return
	}
	api.LogWithUser(username).Info("credit check requested")
	Respond(c, http.StatusOK, gin.H{"response": credits})
}

func (api *API) exportKey(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	keyName := c.Param("name")
	keyNamePrefixed := fmt.Sprintf("%s-%s", username, keyName)
	owns, err := api.um.CheckIfKeyOwnedByUser(username, keyNamePrefixed)
	if err != nil {
		api.LogError(err, eh.KeySearchError)(c, http.StatusBadRequest)
		return
	}
	if !owns {
		api.LogError(errors.New(eh.KeyUseError), eh.KeyUseError)(c, http.StatusBadRequest)
		return
	}
	mnemonic, err := api.ipfs.KeystoreManager.ExportKeyToMnemonic(keyNamePrefixed)
	if err != nil {
		api.LogError(err, eh.KeyExportError)(c, http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": mnemonic})
}
