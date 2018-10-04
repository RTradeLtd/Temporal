package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ChangeAccountPassword is used to change a users password
func (api *API) changeAccountPassword(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)

	oldPassword, exists := c.GetPostForm("old_password")
	if !exists {
		FailWithMissingField(c, "old_password")
		return
	}

	newPassword, exists := c.GetPostForm("new_password")
	if !exists {
		FailWithMissingField(c, "new_password")
		return
	}

	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("password change requested")

	suceeded, err := api.um.ChangePassword(username, oldPassword, newPassword)
	if err != nil {
		api.LogError(err, PasswordChangeError)(c)
		return
	}
	if !suceeded {
		err = fmt.Errorf("password changed failed for user %s to due an unspecified error", username)
		api.LogError(err, PasswordChangeError)(c)
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
	username, exists := c.GetPostForm("username")
	if !exists {
		FailWithMissingField(c, "username")
		return
	}
	password, exists := c.GetPostForm("password")
	if !exists {
		FailWithMissingField(c, "password")
		return
	}
	email, exists := c.GetPostForm("email_address")
	if !exists {
		FailWithMissingField(c, "email_address")
		return
	}
	api.l.WithFields(log.Fields{
		"service": "api",
	}).Info("user account registration detected")

	userModel, err := api.um.NewUserAccount(username, password, email, false)
	if err != nil {
		api.LogError(err, UserAccountCreationError)(c)
		return
	}

	api.l.WithFields(log.Fields{
		"service": "api",
		"user":    username,
	}).Info("user account registered")

	userModel.HashedPassword = "scrubbed"
	Respond(c, http.StatusOK, gin.H{"response": userModel})
}

// CreateIPFSKey is used to create an IPFS key
func (api *API) createIPFSKey(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	keyType, exists := c.GetPostForm("key_type")
	if !exists {
		FailWithMissingField(c, "key_type")
		return
	}

	switch keyType {
	case "rsa":
		break
	case "ed25519":
		break
	default:
		// user error, do not log
		err := fmt.Errorf("%s is invalid key type must be rsa, or ed25519", keyType)
		Fail(c, err, http.StatusBadRequest)
		return
	}

	user, err := api.um.FindByUserName(username)
	if err != nil {
		api.LogError(err, UserSearchError)(c, http.StatusNotFound)
		return
	}
	var cost float64
	// if they haven't made a key before, the first one is free
	if len(user.IPFSKeyIDs) == 0 {
		cost = 0
		err = nil
	} else {
		if keyType == "rsa" {
			cost, err = utils.CalculateAPICallCost("rsa-key", false)
		} else {
			cost, err = utils.CalculateAPICallCost("ed-key", false)
		}
	}
	if err != nil {
		api.LogError(err, CallCostCalculationError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, cost); err != nil && cost > 0 {
		api.LogError(err, InvalidBalanceError)(c, http.StatusPaymentRequired)
		return
	}
	keyBits, exists := c.GetPostForm("key_bits")
	if !exists {
		FailWithMissingField(c, "key_bits")
		api.refundUserCredits(username, "key", cost)
		return
	}

	keyName, exists := c.GetPostForm("key_name")
	if !exists {
		FailWithMissingField(c, "key_name")
		api.refundUserCredits(username, "key", cost)
		return
	}

	keys, err := api.um.GetKeysForUser(username)
	if err != nil {
		api.LogError(err, KeySearchError)(c, http.StatusNotFound)
		api.refundUserCredits(username, "key", cost)
		return
	}
	keyNamePrefixed := fmt.Sprintf("%s-%s", username, keyName)
	for _, v := range keys["key_names"] {
		if v == keyNamePrefixed {
			err = fmt.Errorf("key with name already exists")
			api.LogError(err, DuplicateKeyCreationError)(c, http.StatusConflict)
			api.refundUserCredits(username, "key", cost)
			return
		}
	}
	bitsInt, err := strconv.Atoi(keyBits)
	if err != nil {
		Fail(c, err)
		return
	}
	// if key type is RSA, and size is too small or too large, default to an appropriately size minimum
	if keyType == "rsa" {
		if bitsInt > 4096 || bitsInt < 2048 {
			bitsInt = 2048
		}
	}
	key := queue.IPFSKeyCreation{
		UserName:    username,
		Name:        keyName,
		Type:        keyType,
		Size:        bitsInt,
		NetworkName: "public",
		CreditCost:  cost,
	}

	mqConnectionURL := api.cfg.RabbitMQ.URL

	qm, err := queue.Initialize(queue.IpfsKeyCreationQueue, mqConnectionURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)(c)
		api.refundUserCredits(username, "key", cost)
		return
	}

	if err = qm.PublishMessageWithExchange(key, queue.IpfsKeyExchange); err != nil {
		api.LogError(err, QueuePublishError)(c)
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
		api.LogError(err, KeySearchError)(c)
		return
	}
	// if the user has no keys, fail with an error
	if len(keys["key_names"]) == 0 || len(keys["key_ids"]) == 0 {
		Fail(c, errors.New(NoKeyError), http.StatusNotFound)
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
		api.LogError(err, CreditCheckError)(c, http.StatusBadRequest)
		return
	}
	api.LogWithUser(username).Info("credit check requested")
	Respond(c, http.StatusOK, gin.H{"response": credits})
}
