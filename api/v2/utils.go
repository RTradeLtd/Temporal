package v2

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/RTradeLtd/Temporal/eh"
	"github.com/RTradeLtd/database/v2/models"
	gpaginator "github.com/RTradeLtd/gpaginator"
	"github.com/c2h5oh/datasize"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	jwt "gopkg.in/dgrijalva/jwt-go.v3"
)

const (
	// RtcCostUsd is the price of a single RTC in USD
	RtcCostUsd = 0.125
)

// pageIt is used to serve paginated responses
func (api *API) pageIt(c *gin.Context, db *gorm.DB, model interface{}) {
	page := c.Query("page")
	if page == "" {
		page = "1"
	}
	limit := c.Query("limit")
	if limit == "" {
		limit = "10"
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		Fail(c, err, http.StatusBadRequest)
		return
	}
	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		Fail(c, err, http.StatusBadRequest)
		return
	}
	paged, err := gpaginator.Paging(
		&gpaginator.Param{
			DB:    db,
			Page:  pageInt,
			Limit: limitInt,
			// sort results starting with newest first
			OrderBy: []string{"created_at DESC"},
		},
		model,
	)
	if err != nil {
		api.LogError(c, err, "failed to get paged user upload "+err.Error())(http.StatusBadRequest)
		return
	}
	Respond(c, http.StatusOK, gin.H{"response": paged})
}

// CheckAccessForPrivateNetwork checks if a user has access to a private network
func CheckAccessForPrivateNetwork(username, networkName string, db *gorm.DB) error {
	um := models.NewUserManager(db)
	canUpload, err := um.CheckIfUserHasAccessToNetwork(username, networkName)
	if err != nil {
		return err
	}

	if !canUpload {
		return errors.New("unauthorized access to private network")
	}
	return nil
}

// GetIPFSEndpoint is used to construct the api url to connect to
// for private ipfs networks. in the case of dev mode it returns
// an default, non nexus based ipfs api address
func (api *API) GetIPFSEndpoint(networkName string) string {
	if dev {
		return api.cfg.IPFS.APIConnection.Host + ":" + api.cfg.IPFS.APIConnection.Port
	}
	return api.cfg.Nexus.Host + ":" + api.cfg.Nexus.Delegator.Port + "/network/" + networkName
}

// FileSizeCheck is used to check and validate the size of the uploaded file
func (api *API) FileSizeCheck(size int64) error {
	sizeInt, err := strconv.ParseInt(
		api.cfg.API.SizeLimitInGigaBytes,
		10,
		64,
	)
	if err != nil {
		return err
	}
	gbInt := int64(datasize.GB.Bytes()) * sizeInt
	if size > gbInt {
		return errors.New(eh.FileTooBigError)
	}
	return nil
}

// generateEmailJWTToken is used to generate a jwt token used to validate emails
func (api *API) generateEmailJWTToken(username, verificationString string) (string, error) {
	// generate a jwt with claims to verify email
	verificationJWT := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"user":                    username,
		"emailVerificationString": verificationString,
		"expire":                  time.Now().Add(time.Hour * 24).UTC().String(),
	})
	// return a signed version of the jwt
	return verificationJWT.SignedString([]byte(api.cfg.API.JWT.Key))
}

func (api *API) verifyEmailJWTToken(jwtString, username string) error {
	// parse the jwt for a token
	token, err := jwt.Parse(jwtString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unable to validate signing method: %v", token.Header["alg"])
		} else if method != jwt.SigningMethodHS512 {
			return nil, errors.New("expect hs512 signing method")
		}
		// return byte version of signing key
		return []byte(api.cfg.JWT.Key), nil
	})
	// verify jwt was parsed properly
	if err != nil {
		return err
	}
	// verify that the token is valid
	if !token.Valid {
		return errors.New("failed to validate token")
	}
	// extract claims from token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("failed to parse claims")
	}
	// verify the username matches what we are expected
	if claims["user"] != username {
		return fmt.Errorf("username from claim does not match expected user of %s", username)
	}
	// get user model so we can validate the email verification string
	user, err := api.um.FindByUserName(username)
	if err != nil {
		return errors.New(eh.UserSearchError)
	}
	emailVerificationString, ok := claims["emailVerificationString"].(string)
	if !ok {
		return errors.New("failed to convert verification token to string")
	}
	// validate email verification string
	if claims["emailVerificationString"] != user.EmailVerificationToken {
		return errors.New("failed to validate email verification token")
	}
	// ensure we can cast claims["expire"] to string type
	expireString, ok := claims["expire"].(string)
	if !ok {
		return errors.New("failed to convert expire value to string")
	}
	// parse expire string into time.Time
	expireTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", expireString)
	if err != nil {
		return err
	}
	// validate that the token hasn't expired
	if time.Now().UTC().Unix() > expireTime.Unix() {
		return errors.New("token is expired")
	}
	// enable email activity
	if _, err := api.um.ValidateEmailVerificationToken(username, emailVerificationString); err != nil {
		return err
	}
	return nil
}

// validateUserCredits is used to validate whether or not a user has enough credits to pay for an action
// and if they do, it is deducted from their account
func (api *API) validateUserCredits(username string, cost float64) error {
	availableCredits, err := api.um.GetCreditsForUser(username)
	if err != nil {
		return err
	}
	if availableCredits < cost {
		return errors.New(eh.InvalidBalanceError)
	}
	if _, err := api.um.RemoveCredits(username, cost); err != nil {
		return err
	}
	return nil
}

// refundUserCredits is used to trigger a credit refund for a user, in the event of an API level processing failure.
// Note that we do not do any error handling here, instead we will log the information so that we may manually
// remediate the situation
func (api *API) refundUserCredits(username, callType string, cost float64) {
	if _, err := api.um.AddCredits(username, cost); err != nil {
		api.l.With("user", username, "call_type", callType, "error", err.Error()).Error(eh.CreditRefundError)
	}
}

// validateAdminRequest is used to validate whether or not the requesting user is an administrator
func (api *API) validateAdminRequest(username string) error {
	isAdmin, err := api.um.CheckIfAdmin(username)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New(eh.UnAuthorizedAdminAccess)
	}
	return nil
}

// used to extract needed post forms that should be provided with api calls.
// if the second return parameter, the string is non-empty, this is the name of the field which was missing
// we then use this to fail with a meaningful message
func (api *API) extractPostForms(c *gin.Context, formNames ...string) (map[string]string, string) {
	forms := make(map[string]string)
	for _, name := range formNames {
		value, exists := c.GetPostForm(name)
		if !exists {
			return nil, name
		}
		if value == "" { // prevent empty values counting as valid
			return nil, name
		}
		forms[name] = value
	}
	return forms, ""
}

// ValidateHoldTime is used to perform parsing of requested hold times,
// returning an int64 type of the provided hold time
func (api *API) validateHoldTime(username, holdTime string) (int64, error) {
	var (
		// 1 year
		freeHoldTimeLimitInMonths int64 = 12
		// two years
		nonFreeHoldTimeLimitInMonths int64 = 24
	)
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		return 0, err
	}
	usageTier, err := api.usage.FindByUserName(username)
	if err != nil {
		return 0, err
	}
	if usageTier.Tier == models.Free && holdTimeInt > freeHoldTimeLimitInMonths {
		return 0, errors.New("free accounts are limited to maximum hold times of 12 month")
	} else if usageTier.Tier != models.Free && holdTimeInt > nonFreeHoldTimeLimitInMonths {
		return 0, errors.New("non free accounts are limited to a maximum hold time of 24 months")
	}
	return holdTimeInt, nil
}

func (api *API) ensureLEMaxPinTime(upload *models.Upload, holdTime int64, tier models.DataUsageTier) error {
	var limit time.Time
	switch tier {
	case models.Free:
		limit = time.Now().AddDate(1, 0, 0)
	case models.Paid, models.Partner, models.WhiteLabeled:
		limit = time.Now().AddDate(2, 0, 0)
	default:
		return errors.New("invalid usage tier")
	}
	if upload.GarbageCollectDate.AddDate(0, int(holdTime), 0).After(limit) {
		return errors.New(eh.MaxHoldTimeError)
	}
	return nil
}

func (api *API) getCMCKey() string {
	if os.Getenv("CMC_API") != "" {
		return os.Getenv("CMC_API")
	}
	return api.cmcAPIKey
}

func (api *API) getCaptchaKey() string {
	if os.Getenv("RECAPTCHA_KEY") != "" {
		return os.Getenv("RECAPTCHA_KEY")
	}
	return ""
}
