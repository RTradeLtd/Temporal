package v2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/rtfs"
	"github.com/c2h5oh/datasize"
)

const (
	tooManyCredits = 10.9999997e+07
	testUser       = "testuser"
)

type apiResponse struct {
	Code     int    `json:"code"`
	Response string `json:"response"`
}

// login form structure.
type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type loginResponse struct {
	Expire string `json:"expire"`
	Token  string `json:"token"`
}

func Test_API_Routes_Account(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// setup connection to ipfs-node-1
	im, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		nil,
		time.Minute*10,
	)
	if err != nil {
		t.Fatal(err)
	}
	// create our test api
	testRecorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(testRecorder)
	lensClient := mocks.FakeIndexerAPIClient{}
	api, err := new(cfg, engine, &lensClient, im, false, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	// setup api routes
	if err = api.setupRoutes(); err != nil {
		t.Fatal(err)
	}

	// authenticate with the api to get our token
	// /api/v2/auth/login
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader("{\n  \"username\": \"testuser\",\n  \"password\": \"admin\"\n}"))
	req.Header.Add("Content-Type", "application/json")
	api.r.ServeHTTP(testRecorder, req)
	// validate the http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/auth/login")
	}
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	var loginResp loginResponse
	if err = json.Unmarshal(bodyBytes, &loginResp); err != nil {
		t.Fatal(err)
	}
	// format authorization header
	authHeader := "Bearer " + loginResp.Token

	// verify the username from the token
	// /api/v2/account/token/username
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/account/token/username", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	// validate http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/token/username")
	}
	var apiResp apiResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/token/username")
	}
	if apiResp.Response != "testuser" {
		t.Fatal("bad username recovered from token")
	}

	// get an email verification token
	// /api/v2/account/email/token/get
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/account/email/token/get", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	// validate the http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/email/token/get")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/email/token/get")
	}
	// since we dont have email lets get the token from the database
	um := models.NewUserManager(db)
	user, err := um.FindByUserName("testuser")
	if err != nil {
		t.Fatal(err)
	}
	if user.EmailVerificationToken == "" {
		t.Fatal("failed to set email verification token")
	}

	// verify the email verification token
	// /api/v2/account/email/token/verify
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/account/email/token/verify", nil)
	// setup the call parameters
	req.Header.Add("Authorization", authHeader)
	urlValues := url.Values{}
	urlValues.Add("token", user.EmailVerificationToken)
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/email/token/verify")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/email/token/verify")
	}
	user, err = um.FindByUserName("testuser")
	if err != nil {
		t.Fatal(err)
	}
	if !user.EmailEnabled {
		t.Fatal("failed to enable email address")
	}

	// verify account password change
	// /api/v2/account/password/change
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/account/password/change", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("old_password", "admin")
	urlValues.Add("new_password", "admin1234@")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/password/change")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		fmt.Println(apiResp.Response)
		t.Fatal("bad api status code from /api/v2/account/password/change")
	}

	// create ipfs keys
	// /api/v2/account/key/ipfs/new
	testRecorder = httptest.NewRecorder()
	// test ed25519 keys
	req = httptest.NewRequest("POST", "/api/v2/account/key/ipfs/new", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("key_type", "ed25519")
	urlValues.Add("key_bits", "256")
	urlValues.Add("key_name", "key1")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/key/ipfs/new")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		fmt.Println(apiResp.Response)
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/new")
	}
	testRecorder = httptest.NewRecorder()
	// test rsa keys
	urlValues.Add("key_type", "rsa")
	urlValues.Add("key_bits", "2048")
	urlValues.Add("key_name", "key2")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/key/ipfs/new")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		fmt.Println(apiResp.Response)
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/new")
	}
	// manually create the keys since we arent using queues
	if err = um.AddIPFSKeyForUser("testuser", "key1", "muchwow"); err != nil {
		t.Fatal(err)
	}
	if err = um.AddIPFSKeyForUser("testuser", "key2", "suchkey"); err != nil {
		t.Fatal(err)
	}

	// get ipfs keys
	// /api/v2/account/key/ipfs/get
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/account/key/ipfs/get", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Error("bad http status code from /api/v2/account/key/ipfs/get")
	}
	type keyCreationResponse struct {
		Code     int   `json:"code"`
		Response gin.H `json:"response"`
	}
	var keyCreationResp keyCreationResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &keyCreationResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if keyCreationResp.Code != 200 {
		fmt.Println(apiResp.Response)
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/get")
	}
}

func Test_Utils(t *testing.T) {
	//	t.Skip()
	cfg, err := config.LoadConfig("../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	lensClient := mocks.FakeIndexerAPIClient{}
	api, err := Initialize(cfg, true, &lensClient)
	if err != nil {
		t.Fatal(err)
	}
	if err = api.FileSizeCheck(int64(datasize.GB.Bytes() * 1)); err != nil {
		t.Fatal(err)
	}
	if err = api.FileSizeCheck(int64(datasize.GB.Bytes() * 10)); err == nil {
		t.Fatal("error expected")
	}
	type args struct {
		paymentType string
		blockchain  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ETHFail", args{"ETH", "ETH"}, true},
		{"ETHPass", args{"eth", "ethereum"}, false},
		{"BTCFail", args{"BTC", "BTC"}, true},
		{"BTCPass", args{"btc", "bitcoin"}, false},
		{"LTCFail", args{"LTC", "LTC"}, true},
		{"LTCPass", args{"ltc", "litecoin"}, false},
		{"XMRFail", args{"XMR", "XMR"}, true},
		{"XMRPass", args{"xmr", "monero"}, false},
		{"DASHFail", args{"DASH", "DASH"}, true},
		{"DASHPass", args{"dash", "dash"}, false},
		{"InvalidCoinFail", args{"biiiitcoooonnneeeeeeecccct", "biiiitcoooonnneeeeeeecccct"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := api.getDepositAddress(tt.args.paymentType); (err != nil) != tt.wantErr {
				t.Errorf("getDepositAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if valid := api.validateBlockchain(tt.args.blockchain); !valid != tt.wantErr {
				t.Errorf("validateBlockchain() error = %v, wantErr %v", valid, tt.wantErr)
			}
		})
	}
	if err = api.validateUserCredits(testUser, 1); err != nil {
		t.Fatal(err)
	}
	if err = api.validateUserCredits(testUser, tooManyCredits); err == nil {
		t.Fatal("error expected")
	}
	if err := api.validateAdminRequest(testUser); err != nil {
		t.Fatal(err)
	}
	if err := api.validateAdminRequest("notareallaccount"); err == nil {
		t.Fatal("error expected")
	}
	user, err := api.um.FindByUserName(testUser)
	if err != nil {
		t.Fatal(err)
	}
	previousCreditAmount := user.Credits
	api.refundUserCredits(testUser, "ipfs-pin", 10)
	user, err = api.um.FindByUserName(testUser)
	if user.Credits != previousCreditAmount+10 {
		t.Fatal("failed to refund credits")
	}
	recorder := httptest.NewRecorder()
	testCtx, _ := gin.CreateTestContext(recorder)
	urlValues := url.Values{}
	urlValues.Add("suchkey", "muchvalue")
	testCtx.Request = &http.Request{PostForm: urlValues}
	forms := api.extractPostForms(testCtx, "suchkey")
	if len(forms) == 0 {
		t.Fatal("failed to extract post forms")
	}
	if forms["suchkey"] != "muchvalue" {
		t.Fatal("failed to extract proper postform")
	}
}

func loadDatabase(cfg *config.TemporalConfig) (*gorm.DB, error) {
	return database.OpenDBConnection(database.DBOptions{
		User:           cfg.Database.Username,
		Password:       cfg.Database.Password,
		Address:        cfg.Database.URL,
		Port:           cfg.Database.Port,
		SSLModeDisable: true,
	})
}
