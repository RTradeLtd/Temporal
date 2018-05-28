package api

import (
	"net/http"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/gin-gonic/gin"
)

func RegisterUserAccount(c *gin.Context) {
	ethAddress, exists := c.GetPostForm("eth_address")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eth_address parameter does not exist"})
		return
	}
	password, exists := c.GetPostForm("password")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password parameter does not exist"})
		return
	}
	dbPass := c.MustGet("db_pass").(string)
	db := database.OpenDBConnection(dbPass)
	userManager := models.NewUserManager(db)
	userModel, err := userManager.NewUserAccount(ethAddress, password, false)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userModel.HashedPassword = "scrubbed"
	c.JSON(http.StatusCreated, gin.H{"user": userModel})
	return
}

func RegisterEnterpriseUserAccount(c *gin.Context) {
	ethAddress, exists := c.GetPostForm("eth_address")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "eth_address parameter does not exist"})
		return
	}
	password, exists := c.GetPostForm("password")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password parameter does not exist"})
		return
	}
	dbPass := c.MustGet("db_pass").(string)
	db := database.OpenDBConnection(dbPass)
	userManager := models.NewUserManager(db)
	userModel, err := userManager.NewUserAccount(ethAddress, password, false)
	if err != nil {
		c.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userModel.HashedPassword = "scrubbed"
	c.JSON(http.StatusCreated, gin.H{"user": userModel})
	return
}
