package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ExternalStatusCheck is a systems status check which requires no special permissions
func (api *API) ExternalSystemsCheck(c *gin.Context) {
	Respond(c, http.StatusOK, gin.H{"response": "systems online"})
}
