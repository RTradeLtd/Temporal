package v2

import (
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2/models"
)

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

	// setup fake mock clients
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	fakeWalletService := &mocks.FakeWalletServiceClient{}

	api, _, err := setupAPI(t, fakeLens, fakeOrch, fakeSigner, fakeWalletService, cfg, db)
	if err != nil {
		t.Fatal(err)
	}

	// verify the username from the token
	// /v2/account/token/username
	var apiResp apiResponse
	if err := sendRequest(
		api, "GET", "/v2/account/token/username", 200, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/account/token/username")
	}
	if apiResp.Response != "testuser" {
		t.Fatal("bad username recovered from token")
	}

	// verify account password change - success
	// /v2/account/password/change
	urlValues := url.Values{}
	urlValues.Add("old_password", "admin")
	urlValues.Add("new_password", "admin1234@")
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/v2/account/password/change", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/account/password/change")
	}

	// verify account password change - failure
	// /v2/account/password/change
	urlValues = url.Values{}
	urlValues.Add("old_password", "admin")
	urlValues.Add("new_password", "admin1234@")
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/v2/account/password/change", 400, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 400 {
		t.Fatal("bad api status code from /v2/account/password/change")
	}

	// get ipfs keys - no keys created
	// /v2/account/key/ipfs/get
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "GET", "/v2/account/key/ipfs/get", 404, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 404 {
		t.Fatal("bad api status code from /v2/account/key/ipfs/get")
	}

	// create ipfs keys
	// /v2/account/key/ipfs/new
	urlValues = url.Values{}
	urlValues.Add("key_type", "ed25519")
	urlValues.Add("key_bits", "256")
	urlValues.Add("key_name", "key1")
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/v2/account/key/ipfs/new", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/account/key/ipfs/new")
	}
	// test rsa keys
	urlValues.Add("key_type", "rsa")
	urlValues.Add("key_bits", "2048")
	urlValues.Add("key_name", "key2")
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/v2/account/key/ipfs/new", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/account/key/ipfs/new")
	}
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/account/key/ipfs/new")
	}
	// manually create the keys since we arent using queues
	if err := models.NewUserManager(db).AddIPFSKeyForUser("testuser", "key1", "muchwow"); err != nil {
		t.Fatal(err)
	}
	if err := models.NewUserManager(db).AddIPFSKeyForUser("testuser", "key2", "suchkey"); err != nil {
		t.Fatal(err)
	}

	// create ipfs key - bad key bit
	// /v2/account/key/ipfs/new
	urlValues = url.Values{}
	urlValues.Add("key_type", "ed25519")
	urlValues.Add("key_bits", "notanumber")
	urlValues.Add("key_name", "key1")
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/v2/account/key/ipfs/new", 400, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 400 {
		t.Fatal("bad api status code from /v2/account/key/ipfs/new")
	}

	// get ipfs keys
	// /v2/account/key/ipfs/get
	var mapAPIResp mapAPIResponse
	if err := sendRequest(
		api, "GET", "/v2/account/key/ipfs/get", 200, nil, nil, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api status code from /v2/account/key/ipfs/get")
	}

	// get available credits
	// /v2/account/credits/available
	var floatAPIResp floatAPIResponse
	if err := sendRequest(
		api, "GET", "/v2/account/credits/available", 200, nil, nil, &floatAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if floatAPIResp.Code != 200 {
		t.Fatal("bad api status code from /v2/account/credits/available")
	}

	// test email activation
	// /v2/account/email/verify/:user/:token
	if _, err := api.um.NewUserAccount("verificationtestuser", "password123", "verificationtestuser@example.org"); err != nil {
		t.Fatal(err)
	}
	userModel, err := api.um.GenerateEmailVerificationToken("verificationtestuser")
	if err != nil {
		t.Fatal(err)
	}
	token, err := api.generateEmailJWTToken("verificationtestuser", userModel.EmailVerificationToken)
	if err != nil {
		t.Fatal(err)
	}
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "GET", "/v2/account/email/verify/"+userModel.UserName+"/"+token, 200, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}

	// forgot email
	// /v2/account/email/forgot
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/v2/account/email/forgot", 200, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/account/email/forgot")
	}

	// test@email.com
	// forgot username
	// /v2/forgot/username
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("email_address", "test@email.com")
	if err := sendRequest(
		api, "POST", "/v2/forgot/username", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/forgot/username")
	}

	// test@email.com
	// forgot password
	// /v2/forgot/password
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("email_address", "test@email.com")
	if err := sendRequest(
		api, "POST", "/v2/forgot/password", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/forgot/password")
	}

	// upgrade account
	// /v2/account/upgrade
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/v2/account/upgrade", 200, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/account/upgrade")
	}

	// usage data
	// /v2/account/upgrade
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		api, "GET", "/v2/account/usage", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api status code from /v2/account/usage")
	}
}
