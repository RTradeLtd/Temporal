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
	db := database.OpenDBConnection(dbPass, dbURL)
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
	db := database.OpenDBConnection(dbPass, dbURL)
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

func GetAuthenticatedUserFromContext(c *gin.Context) string {
	claims := jwt.ExtractClaims(c)
	// this is their eth address
	return claims["id"].(string)
}

/*
func CheckForAPIAccess(c *gin.Context) (bool, error) {
	var user models.User
	ethAddress := GetAuthenticatedUserFromContext(c)
	dbPass := c.MustGet("db_pass").(string)
	db := database.OpenDBConnection(dbPass)
	db.Where("eth_address = ?", ethAddress).First(&user)
	if user.CreatedAt == nilTime {
		return false, errors.New("user account does not exist")
	}
	return user.APIAccess, nil
}*/
