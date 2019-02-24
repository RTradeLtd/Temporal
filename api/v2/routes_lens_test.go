package v2

import (
	"errors"
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
	pb "github.com/RTradeLtd/grpc/lensv2"
)

var (
	lensTestHash = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
)

func Test_API_Routes_Lens_Index(t *testing.T) {
	type args struct {
		objectType       string
		objectIdentifier string
		reindex          string
		callCount        int
		err              error
	}
	tests := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{"Index-Requqest-Pass-Reindex", args{"ipld", lensTestHash, "true", 0, nil}, 200},
		{"Index-Request-Pass-NoReindex", args{"ipld", lensTestHash, "false", 0, nil}, 200},
		{"Index-Request-Fail-Reindex", args{"ipld", lensTestHash, "true", 0, errors.New("bad")}, 400},
		{"Index-Request-Fail-NoReindex", args{"ipld", lensTestHash, "false", 0, errors.New("bad")}, 400},
		{"Index-Bad-Object-Type", args{"storj", lensTestHash, "false", 0, errors.New("bad")}, 400},
		{"Index-Bad-Object-Identifier", args{"ipld", "blah", "false", 0, errors.New("bad")}, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			var interfaceAPIResp interfaceAPIResponse
			urlValues := url.Values{}
			urlValues.Add("object_identifier", tt.args.objectIdentifier)
			urlValues.Add("object_type", tt.args.objectType)
			urlValues.Add("reindex", tt.args.reindex)
			fakeLens.IndexReturnsOnCall(tt.args.callCount, &pb.IndexResp{Doc: &pb.Document{Hash: tt.args.objectIdentifier}}, tt.args.err)
			if err := sendRequest(
				api, "POST", "/v2/lens/index", tt.wantStatus, nil, urlValues, &interfaceAPIResp,
			); err != nil {
				t.Fatal(err)
			}
			if interfaceAPIResp.Code != tt.wantStatus {
				t.Fatalf("bad api response status = %v, wantStatus %v", interfaceAPIResp.Code, tt.wantStatus)
			}
		})
	}
}

func Test_API_Routes_Lens_Search(t *testing.T) {
	// index tests
	type args struct {
		keyword   string
		tag       string
		category  string
		mimeType  string
		hash      string
		required  string
		callCount int
		result    bool
		err       error
	}
	testsIndex := []struct {
		name       string
		args       args
		wantStatus int
	}{
		{"Search-Pass-NoResponse", args{"dog", "food", "isnt", "very", "good", "tasting", 0, false, nil}, 400},
		{"Search-Pass-Response", args{"cat", "food", "is", "even", "worse", "tasting", 0, true, nil}, 200},
		{"Search-Fail", args{"pet", "food", "is", "bad", "in", "general", 0, false, errors.New("bad")}, 400},
	}
	for _, tt := range testsIndex {
		t.Run(tt.name, func(t *testing.T) {
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
			if !tt.args.result {
				fakeLens.SearchReturnsOnCall(
					tt.args.callCount,
					&pb.SearchResp{},
					tt.args.err,
				)
			} else {
				fakeLens.SearchReturnsOnCall(
					tt.args.callCount,
					&pb.SearchResp{
						Results: []*pb.SearchResp_Result{
							{
								Score: 0.32,
							},
						},
					},
					tt.args.err,
				)
			}
			var interfaceAPIResp interfaceAPIResponse
			urlValues := url.Values{}
			urlValues.Add("keyword", tt.args.keyword)
			urlValues.Add("tags", tt.args.tag)
			urlValues.Add("categories", tt.args.category)
			urlValues.Add("hashes", tt.args.hash)
			urlValues.Add("required", tt.args.required)
			if err := sendRequest(
				api, "POST", "/v2/lens/search", tt.wantStatus, nil, urlValues, &interfaceAPIResp,
			); err != nil {
				t.Fatal(err)
			}
			if interfaceAPIResp.Code != tt.wantStatus {
				t.Fatalf("bad api response status = %v, wantStatus %v", interfaceAPIResp.Code, tt.wantStatus)
			}
		})
	}
}
