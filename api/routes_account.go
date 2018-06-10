package api

import (
	"net/http"
	"time"

	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
)

var nilTime time.Time

// RegisterUserAccount is used to sign up with temporal and gain web interface access
// you will not be granted API access however, as that needs to be done manually
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
	dbURL := c.MustGet("db_url").(string)
	dbUser := c.MustGet("db_user").(string)
	db, err := database.OpenDBConnection(dbPass, dbURL, dbUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to open connection to database",
		})
		return
	}
	userManager := models.NewUserManager(db)
	userModel, err := userManager.NewUserAccount(ethAddress, password, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userModel.HashedPassword = "scrubbed"
	c.JSON(http.StatusCreated, gin.H{"user": userModel})
	return
}

// RegisterEnterpriseUserAccount is used to register a user account marked as enterprise enabled
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
	dbURL := c.MustGet("db_url").(string)
	dbUser := c.MustGet("db_user").(string)
	db, err := database.OpenDBConnection(dbPass, dbURL, dbUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "unable to open connection to database",
		})
		return
	}
	userManager := models.NewUserManager(db)
	userModel, err := userManager.NewUserAccount(ethAddress, password, false)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userModel.HashedPassword = "scrubbed"
	c.JSON(http.StatusCreated, gin.H{"user": userModel})
	return
}

// GetAuthenticatedUserFromContext is used to pull the eth address of hte user
func GetAuthenticatedUserFromContext(c *gin.Context) string {
	claims := jwt.ExtractClaims(c)
	// this is their eth address
	return claims["id"].(string)
}
