package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SystemsCheck is a basic check of system integrity
func (api *API) SystemsCheck(c *gin.Context) {
	Respond(c, http.StatusOK, gin.H{"response": "systems online"})
}
