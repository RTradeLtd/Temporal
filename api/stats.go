package api

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/gin-gonic/gin"
	stats "github.com/semihalev/gin-stats"
)

func (api *API) getStats(c *gin.Context) {
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(err, eh.NoAPITokenError)(c, http.StatusBadRequest)
		return
	}
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, eh.UnAuthorizedAdminAccess)
		return
	}
	c.JSON(http.StatusOK, stats.Report())
}
