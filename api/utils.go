package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/utils"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var nilTime time.Time

// FilesUploadBucket is the bucket files are stored into before being processed
const FilesUploadBucket = "filesuploadbucket"

// CalculateFileSize helper route used to calculate the size of a file
func CalculateFileSize(c *gin.Context) {
	fileHandler, err := c.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}
	size := utils.CalculateFileSizeInGigaBytes(fileHandler.Size)
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"response": gin.H{
			"file_size_gb":    size,
			"file_size_bytes": fileHandler.Size,
		},
	})
}

// FailNoExistPostForm is a failure used when a post form does not exist
func FailNoExistPostForm(c *gin.Context, formName string) {
	err := fmt.Errorf("%s post form not present", formName)
	c.JSON(http.StatusBadRequest, gin.H{
		"code":     http.StatusBadRequest,
		"response": err.Error(),
	})
}

// FailNotAuthorized is a failure used when a user is unauthorized for an action
func FailNotAuthorized(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"code":     http.StatusForbidden,
		"response": message,
	})
}

// FailOnError is a failure used when an error occurs
func FailOnError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"code":     http.StatusBadRequest,
		"response": err.Error(),
	})
}

// FailOnServerError is an error handler used when a server error occurs
func FailOnServerError(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":     http.StatusInternalServerError,
		"response": err.Error(),
	})
}

// CheckAccessForPrivateNetwork checks if a user has access to a private network
func CheckAccessForPrivateNetwork(ethAddress, networkName string, db *gorm.DB) error {
	um := models.NewUserManager(db)
	canUpload, err := um.CheckIfUserHasAccessToNetwork(ethAddress, networkName)
	if err != nil {
		return err
	}

	if !canUpload {
		return errors.New("unauthorized access to private network")
	}
	return nil
}

// GetAuthenticatedUserFromContext is used to pull the eth address of hte user
func GetAuthenticatedUserFromContext(c *gin.Context) string {
	claims := jwt.ExtractClaims(c)
	// this is their eth address
	return claims["id"].(string)
}
