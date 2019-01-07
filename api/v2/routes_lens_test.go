package v2

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
	pbLensResp "github.com/RTradeLtd/grpc/lens/response"
)

func Test_API_Routes_Lens(t *testing.T) {
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

	// test lens index - missing post form
	// /v2/lens/index
	var apiResp apiResponse
	urlValues := url.Values{}
	urlValues.Add("object_type", "ipld")
	if err := sendRequest(
		api, "POST", "/v2/lens/index", 400, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}

	// test lens index - invalid object type
	// /v2/lens/index
	urlValues = url.Values{}
	urlValues.Add("object_type", "storj")
	urlValues.Add("object_identifier", hash)
	if err := sendRequest(
		api, "POST", "/v2/lens/index", 400, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test lens index - bad format hash
	// /v2/lens/index
	urlValues = url.Values{}
	urlValues.Add("object_type", "ipld")
	urlValues.Add("object_identifier", "notarealipfshash")
	if err := sendRequest(
		api, "POST", "/v2/lens/index", 400, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test lens index - valid object type, no reindex
	// /v2/lens/index
	var mapAPIResp mapAPIResponse
	urlValues = url.Values{}
	urlValues.Add("object_type", "ipld")
	urlValues.Add("object_identifier", hash)
	// setup our mock index response
	fakeLens.IndexReturnsOnCall(0, &pbLensResp.Index{
		Id:       "fakeid",
		Keywords: []string{"protocols", "minivan"},
	}, nil)
	if err := sendRequest(
		api, "POST", "/v2/lens/index", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /v2/lens/index")
	}

	// test lens index - valid object type, with reindex
	// /v2/lens/index
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("object_type", "ipld")
	urlValues.Add("object_identifier", hash)
	urlValues.Add("reindex", "yes")
	// setup our mock index response
	fakeLens.IndexReturnsOnCall(1, &pbLensResp.Index{
		Id:       "fakeid",
		Keywords: []string{"protocols", "minivan"},
	}, nil)
	if err := sendRequest(
		api, "POST", "/v2/lens/index", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /v2/lens/index")
	}

	// test lens index - valid object type, with non yes reindex
	// /v2/lens/index
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("object_type", "ipld")
	urlValues.Add("object_identifier", hash)
	urlValues.Add("reindex", "notyes")
	// setup our mock index response
	fakeLens.IndexReturnsOnCall(2, &pbLensResp.Index{
		Id:       "fakeid",
		Keywords: []string{"protocols", "minivan"},
	}, nil)
	if err := sendRequest(
		api, "POST", "/v2/lens/index", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /v2/lens/index")
	}

	// test lens search - with no objects
	// /v2/lens/search
	apiResp = apiResponse{}
	fakeLens.SearchReturnsOnCall(0, nil, nil)
	urlValues = url.Values{}
	urlValues.Add("keywords", "notarealsearch")
	if err := sendRequest(
		api, "POST", "/v2/lens/search", 400, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	if apiResp.Code != 400 {
		t.Fatal("bad api response code from /v2/lens/search")
	}
	if apiResp.Response != "no results found" {
		t.Fatal("failed to correctly detect no results found")
	}

	// test lens search - with objects
	// /v2/lens/search
	// setup our mock search response
	var lensSearchAPIResp lensSearchAPIResponse
	obj := pbLensResp.Object{
		Name:     hash,
		MimeType: "application/pdf",
		Category: "documents",
	}
	var objs []*pbLensResp.Object
	objs = append(objs, &obj)
	fakeLens.SearchReturnsOnCall(1, &pbLensResp.Results{
		Objects: objs,
	}, nil)
	urlValues = url.Values{}
	urlValues.Add("keywords", "minivan")
	urlValues.Add("keywords", "protocols")
	if err := sendRequest(
		api, "POST", "/v2/lens/search", 200, nil, urlValues, &lensSearchAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	fmt.Println(lensSearchAPIResp)
	if lensSearchAPIResp.Code != 200 {
		t.Fatal("bad api response code from /v2/lens/search")
	}
	if lensSearchAPIResp.Response[0]["name"] != hash {
		t.Fatal("failed to search for correct hash")
	}
	if lensSearchAPIResp.Response[0]["category"] != "documents" {
		t.Fatal("failed to search for correct category")
	}
	if lensSearchAPIResp.Response[0]["mimeType"] != "application/pdf" {
		t.Fatal("failed to search for correct mimetpye")
	}
}
