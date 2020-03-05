package v2

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2/models"
)

func Test_API_Routes_Database(t *testing.T) {
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

	api, err := setupAPI(t, fakeLens, fakeOrch, fakeSigner, fakeWalletService, cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	// create uploads to test searching with
	up1, err := api.upm.NewUpload("testhash123", "file", models.UploadOptions{
		FileName:         "dogpic123.jpg",
		HoldTimeInMonths: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer api.upm.DB.Unscoped().Delete(up1)
	up2, err := api.upm.NewUpload("testhash1234", "file", models.UploadOptions{
		FileName:         "catpic123.jpg",
		HoldTimeInMonths: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer api.upm.DB.Unscoped().Delete(up2)
	up3, err := api.upm.NewUpload("testhash12345", "file", models.UploadOptions{
		FileName:         "bigdogpic123.jpg",
		HoldTimeInMonths: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer api.upm.DB.Unscoped().Delete(up3)

	type args struct {
		wantName  map[string]bool
		query     string
		wantCount int
	}
	tests := []struct {
		name string
		args args
	}{
		{"dog", args{
			map[string]bool{"bigdogpic123": true, "dogpic123": true},
			"%dog%", 2,
		}},
		{"cat", args{
			map[string]bool{"catpic123": true},
			"%cat%", 1,
		}},
		{"pic", args{
			map[string]bool{"dogpic123": true, "bigdogpic123": true, "catpic123": true},
			"%pic%", 3,
		}},
	}
	// test search (non paged)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var interfaceAPIResp interfaceAPIResponse
			if err := sendRequest(
				api, "POST", "/v2/database/uploads/search", 200, nil, url.Values{
					"search_query": []string{tt.args.query},
				}, &interfaceAPIResp,
			); err != nil {
				t.Fatal(err)
			}
			fmt.Printf("DBDEBUG\n%+v\nDBDEBUG\n", interfaceAPIResp)
			/*if len(searchAPIResp.Response) != tt.args.wantCount {
				t.Fatal("bad count")
			}
			for _, up := range searchAPIResp.Response {
				if !tt.args.wantName[up.FileNameLowerCase] {
					t.Fatal("bad upload found")
				}
			}*/
			t.Error("just a test")
		})
	}
	// test search (paged)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var searchAPIResp searchAPIResponse
			if err := sendRequest(
				api, "POST", "/v2/database/uploads/search", 200, nil, url.Values{
					"search_query": []string{tt.args.query},
					"paged":        []string{"true"},
				}, &searchAPIResp,
			); err != nil {
				t.Fatal(err)
			}
			fmt.Printf("DBDEBUG\n%+v\nDBDEBUG\n", searchAPIResp)
			if len(searchAPIResp.Response) != tt.args.wantCount {
				t.Fatal("bad count")
			}
			for _, up := range searchAPIResp.Response {
				if !tt.args.wantName[up.FileNameLowerCase] {
					t.Fatal("bad upload found")
				}
			}
		})
	}

	// test database specific uploads
	// /v2/database/uploads/testuser
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		api, "GET", "/v2/database/uploads", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// test paginated
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequestPaged(
		api,
		"GET",
		"/v2/database/uploads",
		200,
		nil,
		url.Values{"paged": {"true"}},
		&interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// test get encrypted uploads
	// /v2/frontend/uploads/encrypted
	if err := sendRequest(
		api, "GET", "/v2/database/uploads/encrypted", 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequestPaged(
		api,
		"GET",
		"/v2/database/uploads/encrypted",
		200,
		nil,
		url.Values{"paged": {"true"}},
		&interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
}
