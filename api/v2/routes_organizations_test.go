package v2

import (
	"net/http"
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

	// create organization
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v2/org/new", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues := url.Values{}
	urlValues.Add("name", "testorg")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Result().StatusCode != http.StatusOK {
		t.Fatal("bad status returned")
	}
	// find the organization model to ensure it is create
	org, err := models.NewOrgManager(db).FindByName("testorg")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Unscoped().Delete(org)

	// register organization user
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/v2/org/register/user", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("username", "testorg-user")
	urlValues.Add("password", "password123")
	urlValues.Add("email_address", "testorg+22@example.org")
	urlValues.Add("organization_name", "testorg")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Result().StatusCode != http.StatusOK {
		t.Fatal("bad status returned")
	}
	usr, err := models.NewUserManager(db).FindByUserName("testorg-user")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Unscoped().Delete(usr)
	if usr.Organization != "testorg" {
		t.Fatal("bad organization found")
	}
	usg, err := models.NewUsageManager(db).FindByUserName("testorg-user")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Unscoped().Delete(usg)
	if usg.Tier != models.WhiteLabeled {
		t.Fatal("bad tier found")
	}
}
