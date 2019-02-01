package v2

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database/models"
)

func Test_API_Routes_IPFS_Public(t *testing.T) {
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
	// update the users tier
	if err := api.usage.UpdateTier("testuser", models.Plus); err != nil {
		t.Fatal(err)
	}
	// add a file normally
	// /v2/ipfs/public/file/add
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", "../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	fh, err := os.Open("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	if _, err = io.Copy(fileWriter, fh); err != nil {
		t.Fatal(err)
	}
	bodyWriter.Close()
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v2/ipfs/public/file/add", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /v2/ipfs/public/file/add")
	}
	var apiResp apiResponse
	// unmarshal the response
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/ipfs/public/file/add")
	}
	hash = apiResp.Response

	// test pinning - success
	// /v2/ipfs/public/pin
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	if err := sendRequest(
		api, "POST", "/v2/ipfs/public/pin/"+hash, 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from  /v2/ipfs/public/pin")
	}

	// test pinning - failure
	// /v2/ipfs/public/pin
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	if err := sendRequest(
		api, "POST", "/v2/ipfs/public/pin/notarealhash", 400, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 400 {
		t.Fatal("bad api status code from  /v2/ipfs/public/pin")
	}

	// test pubsub publish
	// /v2/ipfs/pubsub/publish/topic
	urlValues = url.Values{}
	urlValues.Add("message", "bar")
	var mapAPIResp mapAPIResponse
	if err := sendRequest(
		api, "POST", "/v2/ipfs/public/pubsub/publish/foo", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api status code from  /v2/pubsub/publish/topic")
	}
	if mapAPIResp.Response["topic"] != "foo" {
		t.Fatal("bad response")
	}
	if mapAPIResp.Response["message"] != "bar" {
		t.Fatal("bad response")
	}

	// test object stat
	// /v2/ipfs/stat
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		api, "GET", "/v2/ipfs/public/stat/"+hash, 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /v2/ipfs/public/stat")
	}

	// test get dag
	// /v2/ipfs/public/dag
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/v2/ipfs/public/dag/"+hash, 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /v2/ipfs/public/dag/")
	}

	// test download
	// /v2/ipfs/utils/download
	if err := sendRequest(
		api, "POST", "/v2/ipfs/utils/download/"+hash, 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test public network beam
	// /v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "public")
	urlValues.Add("destination_network", "public")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		api, "POST", "/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}
}
