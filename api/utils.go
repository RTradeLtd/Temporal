package api

import (
	"fmt"
	"net/http"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func FailNoExist(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": message,
	})
}

func FailNoExistPostForm(c *gin.Context, formName string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": fmt.Sprintf("%s post form not present", formName),
	})
}
func FailNotAuthorized(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": message,
	})
}

func FailOnError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}

func FailedToLoadDatabase(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "failed to load database",
	})
}

func CheckAccessForPrivateNetwork(ethAddress, networkName string, db *gorm.DB) (bool, error) {
	um := models.NewUserManager(db)
	canUpload, err := um.CheckIfUserHasAccessToNetwork(ethAddress, networkName)
	if err != nil {
		return false, err
	}

	if !canUpload {
		return false, nil
	}
	return true, nil
}
