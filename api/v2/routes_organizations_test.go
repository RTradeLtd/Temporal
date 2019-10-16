package v2

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2/models"
)

func Test_API_Routes_Organization(t *testing.T) {
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

	api, testRecorder, err := setupAPI(t, fakeLens, fakeOrch, fakeSigner, fakeWalletService, cfg, db)
	if err != nil {
		t.Fatal(err)
	}

	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v2/org/new", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues := url.Values{}
	urlValues.Add("name", "testorg")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	org, err := models.NewOrgManager(db).FindByName("testorg")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Unscoped().Delete(org)
}
