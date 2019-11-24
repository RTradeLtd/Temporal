package v2

import (
	"errors"
	"net/http"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/database/v2/models"
	"github.com/gin-gonic/gin"
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
	if usage.Tier == models.Free {
		Fail(c, errors.New("free accounts not eligible for ens claim"), http.StatusBadRequest)
		return
	}
	if usage.ClaimedENSName {
		Fail(c, errors.New("user already claimed ens name"), http.StatusBadRequest)
		return
	}
	if err := api.usage.ClaimENSName(username); err != nil {
		api.LogError(c, err, "failed to claim ens name")(http.StatusBadRequest)
		return
	}
	if err := api.queues.ens.PublishMessage(queue.ENSRequest{
		Type:     queue.ENSRegisterSubName,
		UserName: username,
	}); err != nil {
		if err := api.usage.UnclaimENSName(username); err != nil {
			api.l.Errorw("failed to unclaim ens name", "user", username, "error", err)
		}
		api.LogError(c, err, eh.QueuePublishError)(http.StatusBadRequest)
		return
	}
	api.l.Infow("ens name claim request sent to backend", "user", username)
	Respond(c, http.StatusOK, gin.H{
		"response": "request is being processed, you will receive an email with an update when finished",
	})
}
