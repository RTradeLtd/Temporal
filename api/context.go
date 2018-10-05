package api

import (
	"fmt"
	"net/http"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
)

// Fail fails context with given error and optional status code. Defaults to
// http.StatusInternalServerError
func Fail(c *gin.Context, err error, code ...int) {
	c.JSON(status(code), gin.H{
		"code":     status(code),
		"response": err.Error(),
	})
}

// FailWithMessage fails context with given message and optional status code.
// Defaults to http.StatusInternalServerError
func FailWithMessage(c *gin.Context, message string, code ...int) {
	c.JSON(status(code), gin.H{
		"code":     status(code),
		"response": message,
	})
}

// FailWithBadRequest fails context with a bad request error and given message
func FailWithBadRequest(c *gin.Context, message string) {
	FailWithMessage(c, message, http.StatusBadRequest)
}

// FailWithMissingField is a failure used when a post form does not exist
func FailWithMissingField(c *gin.Context, field string) {
	FailWithBadRequest(c, fmt.Sprintf("%s not present", field))
}

// FailNotAuthorized is a failure used when a user is unauthorized for an action
func FailNotAuthorized(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, gin.H{
		"code":     http.StatusForbidden,
		"response": message,
	})
}

// GetAuthenticatedUserFromContext is used to pull the eth address of hte user
func GetAuthenticatedUserFromContext(c *gin.Context) string {
	claims := jwt.ExtractClaims(c)
	// this is their eth address
	return claims["id"].(string)
}

// Respond is a wrapper used to handle API responses
func Respond(c *gin.Context, status int, body gin.H) {
	body["code"] = status
	c.JSON(status, body)
}

// status is used to handle optional status code params
func status(i []int) (status int) {
	if i == nil || len(i) == 0 {
		status = http.StatusInternalServerError
	} else {
		status = i[0]
	}
	return
}
