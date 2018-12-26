package v2

import (
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database/models"
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
	fakeLens := &mocks.FakeIndexerAPIClient{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}

	api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
	if err != nil {
		t.Fatal(err)
	}

	// verify the username from the token
	// /api/v2/account/token/username
	var apiResp apiResponse
	if err := sendRequest(
		api, "GET", "/api/v2/account/token/username", 200, nil, nil, &apiResp,
	); err != nil {
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
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "GET", "/api/v2/account/email/token/get", 200, nil, nil, &apiResp,
	); err != nil {
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
	urlValues := url.Values{}
	urlValues.Add("token", user.EmailVerificationToken)
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/api/v2/account/email/token/verify", 200, nil, urlValues, &apiResp,
	); err != nil {
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
	urlValues = url.Values{}
	urlValues.Add("old_password", "admin")
	urlValues.Add("new_password", "admin1234@")
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/api/v2/account/password/change", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/password/change")
	}

	// create ipfs keys
	// /api/v2/account/key/ipfs/new
	urlValues = url.Values{}
	urlValues.Add("key_type", "ed25519")
	urlValues.Add("key_bits", "256")
	urlValues.Add("key_name", "key1")
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/api/v2/account/key/ipfs/new", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/new")
	}
	// test rsa keys
	urlValues.Add("key_type", "rsa")
	urlValues.Add("key_bits", "2048")
	urlValues.Add("key_name", "key2")
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/api/v2/account/key/ipfs/new", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/new")
	}
	if apiResp.Code != 200 {
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
	var mapAPIResp mapAPIResponse
	if err := sendRequest(
		api, "GET", "/api/v2/account/key/ipfs/get", 200, nil, nil, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/get")
	}

	// get available credits
	// /api/v2/account/credits/available
	var floatAPIResp floatAPIResponse
	if err := sendRequest(
		api, "GET", "/api/v2/account/credits/available", 200, nil, nil, &floatAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if floatAPIResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/credits/available")
	}

	// forgot email
	// /api/v2/account/email/forgot
	apiResp = apiResponse{}
	if err := sendRequest(
		api, "POST", "/api/v2/account/email/forgot", 200, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/email/forgot")
	}

	// test@email.com
	// forgot username
	// /api/v2/forgot/username
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("email_address", "test@email.com")
	if err := sendRequest(
		api, "POST", "/api/v2/forgot/username", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/forgot/username")
	}

	// test@email.com
	// forgot password
	// /api/v2/forgot/password
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("email_address", "test@email.com")
	if err := sendRequest(
		api, "POST", "/api/v2/forgot/password", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/forgot/password")
	}
}
