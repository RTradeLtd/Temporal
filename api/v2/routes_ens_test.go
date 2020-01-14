package v2

import (
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config/v2"
)

func Test_API_Routes_ENS(t *testing.T) {
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
	// instantiate the test api
	api, err := setupAPI(t, fakeLens, fakeOrch, fakeSigner, fakeWalletService, cfg, db)
	if err != nil {
		t.Fatal(err)
	}

	// test update content hash (fail, unclaimed)
	var resp interface{}
	urlValues := url.Values{}
	urlValues.Add("content_hash", hash)
	if err := sendRequest(
		api, "POST", "/v2/ens/update", 400, nil, urlValues, &resp,
	); err != nil {
		t.Fatal(err)
	}

	// test claiming name (200)
	if err := sendRequest(
		api, "POST", "/v2/ens/claim", 200, nil, nil, &resp,
	); err != nil {
		t.Fatal(err)
	}
	// test updating content hash (200)
	urlValues = url.Values{}
	urlValues.Add("content_hash", hash)
	if err := sendRequest(
		api, "POST", "/v2/ens/update", 200, nil, urlValues, &resp,
	); err != nil {
		t.Fatal(err)
	}
}
