package v2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config/v2"
)

func Test_Routes_Swarm(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	cfg.Ethereum.Swarm.URL1 = "http://localhost:8500"
	cfg.Ethereum.Swarm.URL2 = "http://localhost:8500"
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
	authToken := auth(t, api)
	// add a file normally
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
	if _, err := io.Copy(fileWriter, fh); err != nil {
		t.Fatal(err)
	}
	bodyWriter.Close()
	testRecorder := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/v2/swarm/upload", bodyBuf)
	req.Header.Add("Authorization", authToken)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /v2/swarm/upload")
	}
	var apiResp mapAPIResponse
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	if apiResp.Code != 200 {
		t.Fatal("bad api response status code from /v2/swarm/upload")
	}
}

func auth(t *testing.T, api *API) string {
	testRecorder := httptest.NewRecorder()
	req := httptest.NewRequest(
		"POST",
		"/v2/auth/login",
		strings.NewReader(fmt.Sprint("{\n  \"username\": \"testuser2\",\n  \"password\": \"password123!@#$%^&&**(!@#!\"\n}")),
	)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != http.StatusOK {
		t.Fatal("bad status code")
	}
	bodBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	var loginResp loginResponse
	if err = json.Unmarshal(bodBytes, &loginResp); err != nil {
		t.Fatal(err)
	}
	return "Bearer " + loginResp.Token
}
