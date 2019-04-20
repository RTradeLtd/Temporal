package v2

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config/v2"
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
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}

	api, testRecorder, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	// test basic dash payment
	// /v2/payments/create/dash
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v2/payments/dash/create", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues := url.Values{}
	urlValues.Add("credit_value", "10")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test request signed payment message - rtc
	// /v2/payments/request
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/v2/payments/eth/request", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("payment_type", "rtc")
	urlValues.Add("sender_address", "0x0")
	urlValues.Add("credit_value", "10")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test request signed payment message - eth
	// /v2/payments/request
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/v2/payments/eth/request", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("payment_type", "eth")
	urlValues.Add("sender_address", "0x0")
	urlValues.Add("credit_value", "10")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test payment confirmation
	// /v2/payments/confirm
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/v2/payments/eth/confirm", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("payment_number", "10")
	urlValues.Add("tx_hash", "0x1")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
}
