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
	if uploads == nil {
		c.JSON(http.StatusNotFound, nil)
	}
	c.JSON(http.StatusFound, gin.H{"uploads": uploads})
}

func pinHash(c *gin.Context) {
	hash := c.Param("hash")
	err := database.AddHash(c)
	if err != nil {
		return
	}
	manager := rtfs.Initialize()
	err = manager.Shell.Pin(hash)
	if err != nil {
		c.Error(err)
	}
	upload := database.GetUpload(hash, c.PostForm("uploadAddress"))
	c.JSON(http.StatusOK, gin.H{
		"hash":                upload.Hash,
		"uploader":            upload.UploadAddress,
		"hold_time_in_months": upload.HoldTimeInMonths})
}

func addFile(c *gin.Context) {
	fileHandler, err := c.FormFile("file")
	if err != nil {
		c.Error(err)
	}
	openFile, err := fileHandler.Open()
	if err != nil {
		c.Error(err)
	}
	manager := rtfs.Initialize()
	resp, err := manager.Shell.Add(openFile)
	if err != nil {
		c.Error(err)
	}
	database.AddFileHash(c, resp)
	c.JSON(http.StatusOK, gin.H{"response": resp})
}
