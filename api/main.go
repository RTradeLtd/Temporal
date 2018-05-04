package api

import (
	"github.com/RTradeLtd/RTC-IPFS/rtfs"
	"github.com/gin-gonic/gin"
)

// Setup is used to build our routes
func Setup() *gin.Engine {
	r := gin.Default()
	setupRoutes(r)
	return r
}

func setupRoutes(g *gin.Engine) {

	g.POST("/apiv1/add/ipfs-hash/:hash", hashAdd)
}

func hashAdd(c *gin.Context) {
	hash := c.Param("hash")
	manager := rtfs.Initialize()
	err := manager.Shell.Pin(hash)
	if err != nil {
		c.JSON(404, nil)
	}
}
