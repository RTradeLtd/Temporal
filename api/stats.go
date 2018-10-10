package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	stats "github.com/semihalev/gin-stats"
)

func (api *API) getStats(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	if err := api.validateAdminRequest(username); err != nil {
		FailNotAuthorized(c, UnAuthorizedAdminAccess)
		return
	}
	c.JSON(http.StatusOK, stats.Report())
}
