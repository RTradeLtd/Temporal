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
)

func Test_API_Routes_Frontend(t *testing.T) {
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

	// test pin cost calculate
	// /v2/frontend/cost/calculate/:hash/:holTime
	var floatAPIResp floatAPIResponse
	if err := sendRequest(
		api, "GET", "/v2/frontend/cost/calculate/"+hash+"/5", 200, nil, nil, &floatAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if floatAPIResp.Code != 200 {
		t.Fatal("bad response code from /v2/frontend/cost/calculate")
	}

	// test file upload cost calculation
	// /v2/frontend/cost/calculate/file
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
	req := httptest.NewRequest("POST", "/v2/frontend/cost/calculate/file", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /v2/frontend/cost/calculate/file")
	}
	floatAPIResp = floatAPIResponse{}
	// unmarshal the response
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &floatAPIResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if floatAPIResp.Code != 200 {
		t.Fatal("bad api status code from /v2/frontend/cost/calculate/file")
	}
	if floatAPIResp.Response == 0 {
		t.Fatal("failed to calculate cost /v2/frontend/cost/calculate/file")
	}
}
