package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/RTradeLtd/Temporal/models"
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// FilesUploadBucket is the bucket files are stored into before being processed
const FilesUploadBucket = "filesuploadbucket"

// FailNoExist is a failure used when somethign does not exist
func FailNoExist(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": message,
	})
}

// FailNoExistPostForm is a failure used when a post form does not exist
func FailNoExistPostForm(c *gin.Context, formName string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": fmt.Sprintf("%s post form not present", formName),
	})
}

// FailNotAuthorized is a failure used when a user is unauthorized for an action
func FailNotAuthorized(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"error": message,
	})
}

// FailOnError is a failure used when an error occurs
func FailOnError(c *gin.Context, err error) {
	fmt.Println(err)
	c.JSON(http.StatusBadRequest, gin.H{
		"error": err.Error(),
	})
}

// FailedToLoadDatabase is a failure used when the database cant be loaded
func FailedToLoadDatabase(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": "failed to load database",
	})
}

// FailedToLoadMiddleware is a generic failure used whenb middleware cant be loaded
func FailedToLoadMiddleware(c *gin.Context, middlewareName string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error": fmt.Sprintf("failed to load %s middleware", middlewareName),
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
