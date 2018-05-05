package api

import (
	"net/http"
	"os"

	"github.com/RTradeLtd/RTC-IPFS/database"
	"github.com/RTradeLtd/RTC-IPFS/rtfs"
	"github.com/gin-contrib/rollbar"
	"github.com/gin-gonic/gin"
	"github.com/stvp/roll"
)

/*
	roll.Token = "POST_SERVER_ITEM_ACCESS_TOKEN"
	//roll.Environment = "production" // defaults to "development"

	r := gin.Default()
	r.Use(rollbar.Recovery(true))

	r.Run(":8080")
	func l(err error) {
	token := os.Getenv("ROLLBAR_TOKEN")
	rollbar.SetToken(token)
	rollbar.SetServerRoot("github.com/RTradeLtd/RTC-IPFS") // path of project (required for GitHub integration and non-project stacktrace collapsing)

	rollbar.Error(err)

	rollbar.Wait()
}
*/
// Setup is used to build our routes
func Setup() *gin.Engine {
	token := os.Getenv("ROLLBAR_TOKEN")
	roll.Token = token
	roll.Environment = "development"
	r := gin.Default()
	r.Use(rollbar.Recovery(false))
	setupRoutes(r)
	return r
}

func setupRoutes(g *gin.Engine) {

	g.POST("/api/v1/ipfs/pin/:hash", pinHash)
	g.POST("/api/v1/ipfs/add-file", addFile)
	g.GET("/api/v1/ipfs/uploads", getUploads)
}

func getUploads(c *gin.Context) {
	uploads := database.GetUploads()
	c.JSON(http.StatusFound, gin.H{"uploads": uploads})
}

func pinHash(c *gin.Context) {
	hash := c.Param("hash")
	manager := rtfs.Initialize()
	err := manager.Shell.Pin(hash)
	if err != nil {
		c.JSON(404, gin.H{"error": err})
	}
	database.AddHash(c)
	c.JSON(http.StatusOK, gin.H{"hash": hash})
}

func addFile(c *gin.Context) {
	fileHandler, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	openFile, err := fileHandler.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	manager := rtfs.Initialize()
	resp, err := manager.Shell.Add(openFile)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
	database.AddFileHash(resp)
	c.JSON(http.StatusOK, gin.H{"response": resp})
}
