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
	usr, err := api.um.NewUserAccount("ethswarmtest", "password123", "ethswarmtest@google.ca")
	if err != nil {
		t.Fatal(err)
	}
	usr, err = api.um.ValidateEmailVerificationToken("ethswarmtest", usr.EmailVerificationToken)
	if err != nil {
		t.Fatal(err)
	}
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
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /v2/swarm/upload")
	}
	var apiResp apiResponse
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
	if _, err := api.dbm.Upload.FindUploadByHashAndNetwork(apiResp.Response, "etherswarm"); err != nil {
		t.Fatal(err)
	}
}
