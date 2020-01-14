package v2

import (
	"fmt"
	"io/ioutil"
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
	// upgrade account to partner for test.
	if err := api.usage.UpdateTier("testuser", models.Partner); err != nil {
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
	// search for the organization
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/v2/org/get/model", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("name", "testorg")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Result().StatusCode != http.StatusOK {
		t.Fatal("bad status returned")
	}
	// get a billing report
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/v2/org/get/billing/report", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("name", "testorg")
	urlValues.Add("number_of_days", "30")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Result().StatusCode != http.StatusOK {
		t.Fatal("bad status returned")
	}
	// user upload tests

	type args struct {
		name  string
		users []string
		asCSV string
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
	}{
		{"1-user-NoCSV", args{"testorg", []string{"testorg-user"}, "false"}, 200},
		{"2-User-NoCSV-1-Invalid-User", args{"testorg", []string{"testorg-user", "testorg-usernotreal"}, "false"}, 200},
		{"2-User-YesCSV-1-Invalid-User", args{"testorg", []string{"testorg-user", "testorg-usernotreal"}, "true"}, 200},
		{"1-user-YesCSV", args{"testorg", []string{"testorg-user"}, "true"}, 200},
		{"No-OrgName", args{"", []string{"testorg-user"}, "false"}, 400},
		{"No-Users", args{"testorg", nil, "false"}, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRecorder = httptest.NewRecorder()
			req = httptest.NewRequest("POST", "/v2/org/user/uploads", nil)
			req.Header.Add("Authorization", authHeader)
			urlValues = url.Values{}
			// set org name
			urlValues.Add("name", tt.args.name)
			// set users
			for _, user := range tt.args.users {
				urlValues.Add("users", user)
			}
			// set as csv
			urlValues.Add("as_csv", tt.args.asCSV)
			req.PostForm = urlValues
			api.r.ServeHTTP(testRecorder, req)
			if testRecorder.Result().StatusCode != tt.wantCode {
				data, err := ioutil.ReadAll(testRecorder.Result().Body)
				if err != nil {
					t.Error(err)
				}
				fmt.Printf("response:\t%s\t", string(data))
				t.Fatal("bad status")
			}
		})
	}
}
