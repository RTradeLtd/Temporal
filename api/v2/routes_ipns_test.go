package v2

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database/models"
	shell "github.com/RTradeLtd/go-ipfs-api"
)

var (
	validIPNSTestPath   = "/ipns/docs.api.temporal.cloud"
	invalidIPNSTestPath = "/snpi/docs.api.temporal.cloud"
	badIPNSTestPath     = "/not/a/real/path"
	validResolveResult  = "/ipfs/QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
)

func Test_API_Routes_IPNS_Publish(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// create a fake key for testing purposes
	models.NewUserManager(db).AddIPFSKeyForUser("testuser", "mytestkey", "suchkeymuchwow")
	type args struct {
		hash     string
		lifeTime string
		ttl      string
		key      string
		resolve  string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{"Fail-Does-Not-Own-Key", args{hash, "24h", "1h", "notarealkeythisuserowns", "true"}, 400},
		{"Fail-Bad-Hash", args{"notavalidipfshash", "24h", "1h", "mytestkey", "true"}, 400},
		{"Success", args{hash, "24h", "1h", "mytestkey", "true"}, 200},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup fake mock clients
			fakeLens := &mocks.FakeLensV2Client{}
			fakeOrch := &mocks.FakeServiceClient{}
			fakeSigner := &mocks.FakeSignerClient{}

			api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
			if err != nil {
				t.Fatal(err)
			}
			var apiResp apiResponse
			urlValues := url.Values{}
			urlValues.Add("hash", tt.args.hash)
			urlValues.Add("life_time", tt.args.lifeTime)
			urlValues.Add("ttl", tt.args.ttl)
			urlValues.Add("key", tt.args.key)
			urlValues.Add("resolve", tt.args.resolve)
			if err := sendRequest(
				api, "POST", "/v2/ipns/public/publish/details", tt.wantStatus, nil, urlValues, &apiResp,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_API_Routes_IPNS(t *testing.T) {
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

	um := models.NewUserManager(db)
	um.AddIPFSKeyForUser("testuser", "mytestkey", "suchkeymuchwow")

	// test get ipns records
	// /v2/ipns/records
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/v2/ipns/records", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /v2/ipns/records")
	}

	// test get ipns records
	// /v2/ipns/records
	// spoof a fake record as we arent using the queues in this test
	var ipnsAPIResp ipnsAPIResponse
	im := models.NewIPNSManager(db)
	if _, err := im.CreateEntry("fakekey", "fakehash", "fakekeyname", "public", "testuser", time.Minute, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := sendRequest(
		api, "GET", "/v2/ipns/records", 200, nil, nil, &ipnsAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if ipnsAPIResp.Code != 200 {
		t.Fatal("bad api status code from /v2/ipns/private/publish/details")
	}
	if len(*ipnsAPIResp.Response) == 0 {
		t.Fatal("no records discovered")
	}

}

func Test_API_Routes_IPNS_Pin(t *testing.T) {

	type args struct {
		holdTime string
		ipnsPath string
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{"Success", args{"1", validIPNSTestPath}, 200},
		{"Fail-Bad-Hold-Time", args{"notanumber", validIPNSTestPath}, 400},
		{"Fail-Bad-IPNS-Path", args{"1", badIPNSTestPath}, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
			if err != nil {
				t.Fatal(err)
			}
			fakeManager := &mocks.FakeManager{}
			api.ipfs = fakeManager
			fakeManager.ResolveReturnsOnCall(0, validResolveResult, nil)
			fakeManager.StatReturnsOnCall(0, &shell.ObjectStats{CumulativeSize: 5000000}, nil)
			fakeManager.StatReturnsOnCall(1, &shell.ObjectStats{CumulativeSize: 5000000}, nil)
			var apiResp apiResponse
			urlValues := url.Values{}
			urlValues.Add("hold_time", tt.args.holdTime)
			urlValues.Add("ipns_path", tt.args.ipnsPath)
			if err := sendRequest(
				api, "POST", "/v2/ipns/public/pin", tt.wantStatus, nil, urlValues, &apiResp,
			); err != nil {
				t.Fatal(err)
			}
		})
	}
}
