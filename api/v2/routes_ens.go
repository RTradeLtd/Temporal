package v2

import (
	"errors"
	"net/http"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/database/v2/models"
	"github.com/gin-gonic/gin"
	gocid "github.com/ipfs/go-cid"
)

// ClaimENSName is used to claim a username based ens subdomain
func (api *API) ClaimENSName(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	usage, err := api.usage.FindByUserName(username)
	if err != nil {
		api.LogError(c, err, eh.UserSearchError)(http.StatusBadRequest)
		return
	}
	// prevent processing if account is free tier
	if usage.Tier == models.Free {
		Fail(c, errors.New("free accounts not eligible for ens claim"), http.StatusBadRequest)
		return
	}
	// prevent processing if account is using an already claimed name
	if usage.ClaimedENSName {
		Fail(c, errors.New("user already claimed ens name"), http.StatusBadRequest)
		return
	}
	if err := api.validateUserCredits(username, 0.15); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	// mark account as having claimed ens name
	if err := api.usage.ClaimENSName(username); err != nil {
		api.LogError(c, err, "failed to claim ens name")(http.StatusBadRequest)
		api.refundUserCredits(username, "ens", 0.15)
		return
	}
	if err := api.queues.ens.PublishMessage(queue.ENSRequest{
		Type:     queue.ENSRegisterSubName,
		UserName: username,
	}); err != nil {
		// this fails, unclaim
		if err := api.usage.UnclaimENSName(username); err != nil {
			api.l.Errorw("failed to unclaim ens name", "user", username, "error", err)
		}
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		api.refundUserCredits(username, "ens", 0.15)
		return
	}
	api.l.Infow("ens name claim request sent to backend", "user", username)
	Respond(c, http.StatusOK, gin.H{
		"response": "request is being processed, you will receive an email with an update when finished",
	})
}

// UpdateContentHash is used to update the content hash for an ens name
func (api *API) UpdateContentHash(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	usage, err := api.usage.FindByUserName(username)
	if err != nil {
		api.LogError(c, err, eh.UserSearchError)(http.StatusBadRequest)
		return
	}
	// prevent processing if account is free tier
	if usage.Tier == models.Free {
		Fail(c, errors.New("free accounts not eligible for ens claim"), http.StatusBadRequest)
		return
	}
	// prevent processing if name is unclaimed
	if !usage.ClaimedENSName {
		Fail(c, errors.New("user has not claimed ens name"), http.StatusBadRequest)
		return
	}
	// extract post forms
	forms, missingField := api.extractPostForms(c, "content_hash")
	if missingField != "" {
		FailWithMissingField(c, missingField)
		return
	}
	// validate the content is is valid
	if _, err := gocid.Decode(forms["content_hash"]); err != nil {
		Fail(c, err)
		return
	}
	if err := api.validateUserCredits(username, 0.15); err != nil {
		api.LogError(c, err, eh.InvalidBalanceError)(http.StatusPaymentRequired)
		return
	}
	if err := api.queues.ens.PublishMessage(queue.ENSRequest{
		Type:        queue.ENSUpdateContentHash,
		UserName:    username,
		ContentHash: forms["content_hash"],
	}); err != nil {
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		api.refundUserCredits(username, "ens", 0.15)
		return
	}
	api.l.Infow("ens content hash update request sent to backend", "user", username)
	Respond(c, http.StatusOK, gin.H{
		"response": "request is being processed, you will receive an email with an update when finished",
	})
}
