package v2

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
)

func Test_API_Routes_Payments(t *testing.T) {
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

	api, testRecorder, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	// test basic dash payment
	// /api/v2/payments/create/dash
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v2/payments/create/dash", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues := url.Values{}
	urlValues.Add("credit_value", "10")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test request signed payment message - rtc
	// /api/v2/payments/request
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/payments/request", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("payment_type", "rtc")
	urlValues.Add("sender_address", "0x0")
	urlValues.Add("credit_value", "10")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test request signed payment message - eth
	// /api/v2/payments/request
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/payments/request", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("payment_type", "eth")
	urlValues.Add("sender_address", "0x0")
	urlValues.Add("credit_value", "10")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test payment confirmation
	// /api/v2/payments/confirm
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/payments/confirm", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("payment_number", "10")
	urlValues.Add("tx_hash", "0x1")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test valid deposit address
	args := []string{"eth", "rtc", "btc", "ltc", "xmr", "dash"}
	for _, v := range args {
		if err := sendRequest(
			api, "GET", "/api/v2/payments/deposit/address/"+v, 200, nil, nil, nil,
		); err != nil {
			t.Fatal(err)
		}
	}

	// test invalid deposit address
	if err := sendRequest(
		api, "GET", "/api/v2/payments/deposit/address/invalidType", 400, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}
}
