package api

import "github.com/gin-gonic/gin"

func CreateIPFSNetworkEntryInDatabase(c *gin.Context) {
	// lock down as admin route for now
	ethAddress := GetAuthenticatedUserFromContext(c)
	if ethAddress != AdminAddress {
		FailNotAuthorized(c, "unauthorized access")
	}
}
