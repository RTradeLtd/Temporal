package v2

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

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
func GetAuthenticatedUserFromContext(c *gin.Context) (string, error) {
	claims := jwt.ExtractClaims(c)
	id, ok := claims["id"]
	if !ok {
		return "", errors.New("failed to extract claim id")
	}
	strID, ok := id.(string)
	if !ok {
		return "", errors.New("failed to parse claim id")
	}
	if strID == "" {
		return "", errors.New("no username recovered")
	}
	// this is their eth address
	return strID, nil
}

// GetAuthToken is used to retrieve the jwt token
// from an authenticated request
func GetAuthToken(c *gin.Context) string {
	return strings.Split(c.GetHeader("Authorization"), ":")[1]
}

// Respond is a wrapper used to handle API responses
func Respond(c *gin.Context, status int, body gin.H) {
	body["code"] = status
	c.JSON(status, body)
}

// status is used to handle optional status code params
func status(i []int) (status int) {
	if i == nil || len(i) == 0 {
		status = http.StatusBadRequest
	} else {
		status = i[0]
	}
	return
}
