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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/rtfs"
	"github.com/c2h5oh/datasize"
)

const (
	tooManyCredits = 10.9999997e+07
	testUser       = "testuser"
)

var (
	hash = "QmPY5iMFjNZKxRbUZZC85wXb9CFgNSyzAy1LxwL62D8VGr"
)

type apiResponse struct {
	Code     int    `json:"code"`
	Response string `json:"response"`
}

// login form structure.
type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type loginResponse struct {
	Expire string `json:"expire"`
	Token  string `json:"token"`
}

func Test_API_Routes_Misc(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	// setup connection to ipfs-node-1
	im, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		nil,
		time.Minute*10,
	)
	if err != nil {
		t.Fatal(err)
	}
	// create our test api
	testRecorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(testRecorder)
	lensClient := mocks.FakeIndexerAPIClient{}
	api, err := new(cfg, engine, &lensClient, im, false, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	// setup api routes
	if err = api.setupRoutes(); err != nil {
		t.Fatal(err)
	}

	// authenticate with the the api to get our token for testing
	// /api/v2/auth/login
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader("{\n  \"username\": \"testuser\",\n  \"password\": \"admin\"\n}"))
	req.Header.Add("Content-Type", "application/json")
	api.r.ServeHTTP(testRecorder, req)
	// validate the http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/auth/login")
	}
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	var loginResp loginResponse
	if err = json.Unmarshal(bodyBytes, &loginResp); err != nil {
		t.Fatal(err)
	}
	// format authorization header
	authHeader := "Bearer " + loginResp.Token

	// test systems check route
	// //api/v2/systems/check
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/systems/check", nil)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/systems/check")
	}
	var apiResp apiResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/systems/check")
	}
	if apiResp.Response != "systems online" {
		t.Fatal("bad system status recovered")
	}

	// test systems statistics
	// /api/v2/statistics/stats
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/statistics/stats", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/statistics/stats")
	}

	// test mini create bucket
	// /api/v2/admin/mini/create/bucket
	if err = os.Setenv("MINI_SSL_ENABLE", "false"); err != nil {
		t.Fatal(err)
	}
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/admin/mini/create/bucket", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues := url.Values{}
	urlValues.Add("bucket_name", "filesuploadbucket")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/admin/mini/create/bucket")
	}
}

func Test_API_Routes_IPFS(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	// setup connection to ipfs-node-1
	im, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		nil,
		time.Minute*10,
	)
	if err != nil {
		t.Fatal(err)
	}
	// create our test api
	testRecorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(testRecorder)
	lensClient := mocks.FakeIndexerAPIClient{}
	api, err := new(cfg, engine, &lensClient, im, false, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	// setup api routes
	if err = api.setupRoutes(); err != nil {
		t.Fatal(err)
	}

	// authenticate with the the api to get our token for testing
	// /api/v2/auth/login
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader("{\n  \"username\": \"testuser\",\n  \"password\": \"admin\"\n}"))
	req.Header.Add("Content-Type", "application/json")
	api.r.ServeHTTP(testRecorder, req)
	// validate the http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/auth/login")
	}
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	var loginResp loginResponse
	if err = json.Unmarshal(bodyBytes, &loginResp); err != nil {
		t.Fatal(err)
	}
	// format authorization header
	authHeader := "Bearer " + loginResp.Token

	// add a file normally
	// /api/v2/ipfs/public/file/add
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", "../../testenv/config.json")
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
	req = httptest.NewRequest("POST", "/api/v2/ipfs/public/file/add", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /api/v2/ipfs/public/file/add")
	}
	var apiResp apiResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipfs/public/file/add")
	}
	hash = apiResp.Response

	// add a file advanced
	// /api/v2/ipfs/public/file/add/advanced
	bodyBuf = &bytes.Buffer{}
	bodyWriter = multipart.NewWriter(bodyBuf)
	fileWriter, err = bodyWriter.CreateFormFile("file", "../../testenv/config.json")
	fh, err = os.Open("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()
	if _, err = io.Copy(fileWriter, fh); err != nil {
		t.Fatal(err)
	}
	bodyWriter.Close()
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/ipfs/public/file/add/advanced", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	urlValues.Add("passphrase", "password123")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /api/v2/ipfs/public/file/add/advanced")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipfs/public/file/add/advanced")
	}

	// test pinning
	// /api/v2/ipfs/public/pin
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/ipfs/public/pin/"+hash, nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	// validate the http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from  /api/v2/ipfs/public/pin")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/ipfs/public/pin")
	}

	// test pin check
	// /api/v2/ipfs/public/check/pin
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/ipfs/public/pin/check/"+hash, nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	// validate the http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/ipfs/public/check/pin")
	}
	type pinCheckResponse struct {
		Code     int  `json:"code"`
		Response bool `json:"response"`
	}
	var pinCheckResp pinCheckResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &pinCheckResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if pinCheckResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/ipfs/public/check/pin")
	}

	// test pubsub publish
	// /api/v2/ipfs/pubsub/publish/topic
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/ipfs/public/pubsub/publish/foo", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("message", "bar")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/pubsub/publish/topic")
	}
	type pubSubResponse struct {
		Code     int   `json:"code"`
		Response gin.H `json:"response"`
	}
	var pubSubResp pubSubResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &pubSubResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if pubSubResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/pubsub/publish/topic")
	}
	if pubSubResp.Response["topic"] != "foo" {
		t.Fatal("bad response")
	}
	if pubSubResp.Response["message"] != "bar" {
		t.Fatal("bad response")
	}

	// test object stat
	// /api/v2/ipfs/stat
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/ipfs/public/stat/"+hash, nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/ipfs/public/stat/")
	}
	type statResponse struct {
		Code     int         `json:"code"`
		Response interface{} `json:"response"`
	}
	var statResp statResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &statResp); err != nil {
		t.Fatal(err)
	}
	if statResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/public/stat")
	}

	// test get dag
	// /api/v2/ipfs/public/dag
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/ipfs/public/dag/"+hash, nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/ipfs/public/dag/")
	}
	type dagResponse struct {
		Code     int         `json:"code"`
		Response interface{} `json:"response"`
	}
	var dagResp dagResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &dagResp); err != nil {
		t.Fatal(err)
	}
	if dagResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/public/dag/")
	}

	// test download
	// /api/v2/ipfs/public/download
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/ipfs/public/download/"+hash, nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		fmt.Println(testRecorder.Code)
		t.Fatal("bad http status code from/api/v2/ipfs/public/download")
	}
}

func Test_API_Routes_Frontend(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	// setup connection to ipfs-node-1
	im, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		nil,
		time.Minute*10,
	)
	if err != nil {
		t.Fatal(err)
	}
	// create our test api
	testRecorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(testRecorder)
	lensClient := mocks.FakeIndexerAPIClient{}
	api, err := new(cfg, engine, &lensClient, im, false, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	// setup api routes
	if err = api.setupRoutes(); err != nil {
		t.Fatal(err)
	}

	// authenticate with the the api to get our token for testing
	// /api/v2/auth/login
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader("{\n  \"username\": \"testuser\",\n  \"password\": \"admin\"\n}"))
	req.Header.Add("Content-Type", "application/json")
	api.r.ServeHTTP(testRecorder, req)
	// validate the http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/auth/login")
	}
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	var loginResp loginResponse
	if err = json.Unmarshal(bodyBytes, &loginResp); err != nil {
		t.Fatal(err)
	}
	// format authorization header
	authHeader := "Bearer " + loginResp.Token

	// test get encrypted uploads
	// /api/v2/frontend/uploads/encrypted
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/frontend/uploads/encrypted", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/frontend/uploads/encrypted")
	}

	// test pin cost calculate
	// /api/v2/frontend/cost/calculate/:hash/:holTime
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/frontend/cost/calculate/"+hash+"/5", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/frontend/cost/calculate")
	}
	type costCalculateResponse struct {
		Code     int     `json:"code"`
		Response float64 `json:"response"`
	}
	var costCalculateResp costCalculateResponse
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &costCalculateResp); err != nil {
		t.Fatal(err)
	}
	if costCalculateResp.Code != 200 {
		t.Fatal("bad response code from /api/v2/frontend/cost/calculate")
	}
	if costCalculateResp.Response == 0 {
		t.Fatal("failed to calculate cost /api/v2/frontend/cost/calculate")
	}

	// test file upload cost calculation
	// /api/v2/frontend/cost/calculate/file
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	fileWriter, err := bodyWriter.CreateFormFile("file", "../../testenv/config.json")
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
	req = httptest.NewRequest("POST", "/api/v2/frontend/cost/calculate/file", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /api/v2/frontend/cost/calculate/file")
	}
	costCalculateResp = costCalculateResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &costCalculateResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if costCalculateResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/frontend/cost/calculate/file")
	}
	if costCalculateResp.Response == 0 {
		t.Fatal("failed to calculate cost /api/v2/frontend/cost/calculate/file")
	}
}

func Test_API_Routes_Database(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	// setup connection to ipfs-node-1
	im, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		nil,
		time.Minute*10,
	)
	if err != nil {
		t.Fatal(err)
	}
	// create our test api
	testRecorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(testRecorder)
	lensClient := mocks.FakeIndexerAPIClient{}
	api, err := new(cfg, engine, &lensClient, im, false, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	// setup api routes
	if err = api.setupRoutes(); err != nil {
		t.Fatal(err)
	}

	// authenticate with the the api to get our token for testing
	// /api/v2/auth/login
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader("{\n  \"username\": \"testuser\",\n  \"password\": \"admin\"\n}"))
	req.Header.Add("Content-Type", "application/json")
	api.r.ServeHTTP(testRecorder, req)
	// validate the http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/auth/login")
	}
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	var loginResp loginResponse
	if err = json.Unmarshal(bodyBytes, &loginResp); err != nil {
		t.Fatal(err)
	}
	// format authorization header
	authHeader := "Bearer " + loginResp.Token

	// test database global uploads
	// /api/v2/database/uploads
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/database/uploads", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/database/uploads")
	}
	type uploadsResponse struct {
		Code     int         `json:"code"`
		Response interface{} `json:"response"`
	}
	var uploadsResp uploadsResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &uploadsResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if uploadsResp.Code != 200 {
		t.Fatal("bad api status code from api/v2/database/uploads")
	}

	// test database specific uploads
	// /api/v2/database/uploads/testuser
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/database/uploads/testuser", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/database/uploads/testuser")
	}
	uploadsResp = uploadsResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &uploadsResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if uploadsResp.Code != 200 {
		t.Fatal("bad api status code from api/v2/database/uploads/testuser")
	}
}

func Test_API_Routes_Account(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// setup connection to ipfs-node-1
	im, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		nil,
		time.Minute*10,
	)
	if err != nil {
		t.Fatal(err)
	}
	// create our test api
	testRecorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(testRecorder)
	lensClient := mocks.FakeIndexerAPIClient{}
	api, err := new(cfg, engine, &lensClient, im, false, os.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	// setup api routes
	if err = api.setupRoutes(); err != nil {
		t.Fatal(err)
	}

	// authenticate with the the api to get our token for testing
	// /api/v2/auth/login
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader("{\n  \"username\": \"testuser\",\n  \"password\": \"admin\"\n}"))
	req.Header.Add("Content-Type", "application/json")
	api.r.ServeHTTP(testRecorder, req)
	// validate the http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/auth/login")
	}
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	var loginResp loginResponse
	if err = json.Unmarshal(bodyBytes, &loginResp); err != nil {
		t.Fatal(err)
	}
	// format authorization header
	authHeader := "Bearer " + loginResp.Token

	// verify the username from the token
	// /api/v2/account/token/username
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/account/token/username", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	// validate http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/token/username")
	}
	var apiResp apiResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/token/username")
	}
	if apiResp.Response != "testuser" {
		t.Fatal("bad username recovered from token")
	}

	// get an email verification token
	// /api/v2/account/email/token/get
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/account/email/token/get", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	// validate the http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/email/token/get")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/email/token/get")
	}
	// since we dont have email lets get the token from the database
	um := models.NewUserManager(db)
	user, err := um.FindByUserName("testuser")
	if err != nil {
		t.Fatal(err)
	}
	if user.EmailVerificationToken == "" {
		t.Fatal("failed to set email verification token")
	}

	// verify the email verification token
	// /api/v2/account/email/token/verify
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/account/email/token/verify", nil)
	// setup the call parameters
	req.Header.Add("Authorization", authHeader)
	urlValues := url.Values{}
	urlValues.Add("token", user.EmailVerificationToken)
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/email/token/verify")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/email/token/verify")
	}
	user, err = um.FindByUserName("testuser")
	if err != nil {
		t.Fatal(err)
	}
	if !user.EmailEnabled {
		t.Fatal("failed to enable email address")
	}

	// verify account password change
	// /api/v2/account/password/change
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/account/password/change", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("old_password", "admin")
	urlValues.Add("new_password", "admin1234@")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/password/change")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		fmt.Println(apiResp.Response)
		t.Fatal("bad api status code from /api/v2/account/password/change")
	}

	// create ipfs keys
	// /api/v2/account/key/ipfs/new
	testRecorder = httptest.NewRecorder()
	// test ed25519 keys
	req = httptest.NewRequest("POST", "/api/v2/account/key/ipfs/new", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("key_type", "ed25519")
	urlValues.Add("key_bits", "256")
	urlValues.Add("key_name", "key1")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/key/ipfs/new")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		fmt.Println(apiResp.Response)
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/new")
	}
	testRecorder = httptest.NewRecorder()
	// test rsa keys
	urlValues.Add("key_type", "rsa")
	urlValues.Add("key_bits", "2048")
	urlValues.Add("key_name", "key2")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/key/ipfs/new")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		fmt.Println(apiResp.Response)
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/new")
	}
	// manually create the keys since we arent using queues
	if err = um.AddIPFSKeyForUser("testuser", "key1", "muchwow"); err != nil {
		t.Fatal(err)
	}
	if err = um.AddIPFSKeyForUser("testuser", "key2", "suchkey"); err != nil {
		t.Fatal(err)
	}

	// get ipfs keys
	// /api/v2/account/key/ipfs/get
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/account/key/ipfs/get", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Error("bad http status code from /api/v2/account/key/ipfs/get")
	}
	type keyCreationResponse struct {
		Code     int   `json:"code"`
		Response gin.H `json:"response"`
	}
	var keyCreationResp keyCreationResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &keyCreationResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if keyCreationResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/get")
	}

	// get available credits
	// /api/v2/account/credits/available
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/api/v2/account/credits/available", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Error("bad http status code from /api/v2/account/credits/available")
	}
	type creditAPIResponse struct {
		Code     int     `json:"code"`
		Response float64 `json:"response"`
	}
	var creditResp creditAPIResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &creditResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/credits/available")
	}

	// get available credits
	// /api/v2/auth/register
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/auth/register", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("username", "testuser2")
	urlValues.Add("password", "password123")
	urlValues.Add("email_address", "testuser2+test@example.org")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Error("bad http status code from /api/v2/auth/register")
	}
	type signupResponse struct {
		Code     int         `json:"code"`
		Response models.User `json:"response"`
	}
	var signupResp signupResponse
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &signupResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if signupResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/auth/register")
	}

	// forgot email
	// /api/v2/account/email/forgot
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/account/email/forgot", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/account/email/forgot")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/email/forgot")
	}

	// test@email.com
	// forgot username
	// /api/v2/forgot/username
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/forgot/username", nil)
	urlValues = url.Values{}
	urlValues.Add("email_address", "test@email.com")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/forgot/username")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/forgot/username")
	}

	// test@email.com
	// forgot password
	// /api/v2/forgot/password
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/forgot/password", nil)
	urlValues = url.Values{}
	urlValues.Add("email_address", "test@email.com")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/forgot/password")
	}
	apiResp = apiResponse{}
	// unmarshal the response
	bodyBytes, err = ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(bodyBytes, &apiResp); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/forgot/password")
	}
}
func Test_Utils(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	lensClient := mocks.FakeIndexerAPIClient{}
	api, err := Initialize(cfg, true, &lensClient)
	if err != nil {
		t.Fatal(err)
	}
	if err = api.FileSizeCheck(int64(datasize.GB.Bytes() * 1)); err != nil {
		t.Fatal(err)
	}
	if err = api.FileSizeCheck(int64(datasize.GB.Bytes() * 10)); err == nil {
		t.Fatal("error expected")
	}
	type args struct {
		paymentType string
		blockchain  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"ETHFail", args{"ETH", "ETH"}, true},
		{"ETHPass", args{"eth", "ethereum"}, false},
		{"BTCFail", args{"BTC", "BTC"}, true},
		{"BTCPass", args{"btc", "bitcoin"}, false},
		{"LTCFail", args{"LTC", "LTC"}, true},
		{"LTCPass", args{"ltc", "litecoin"}, false},
		{"XMRFail", args{"XMR", "XMR"}, true},
		{"XMRPass", args{"xmr", "monero"}, false},
		{"DASHFail", args{"DASH", "DASH"}, true},
		{"DASHPass", args{"dash", "dash"}, false},
		{"InvalidCoinFail", args{"biiiitcoooonnneeeeeeecccct", "biiiitcoooonnneeeeeeecccct"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := api.getDepositAddress(tt.args.paymentType); (err != nil) != tt.wantErr {
				t.Errorf("getDepositAddress() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if valid := api.validateBlockchain(tt.args.blockchain); !valid != tt.wantErr {
				t.Errorf("validateBlockchain() error = %v, wantErr %v", valid, tt.wantErr)
			}
		})
	}
	if err = api.validateUserCredits(testUser, 1); err != nil {
		t.Fatal(err)
	}
	if err = api.validateUserCredits(testUser, tooManyCredits); err == nil {
		t.Fatal("error expected")
	}
	if err := api.validateAdminRequest(testUser); err != nil {
		t.Fatal(err)
	}
	if err := api.validateAdminRequest("notareallaccount"); err == nil {
		t.Fatal("error expected")
	}
	user, err := api.um.FindByUserName(testUser)
	if err != nil {
		t.Fatal(err)
	}
	previousCreditAmount := user.Credits
	api.refundUserCredits(testUser, "ipfs-pin", 10)
	user, err = api.um.FindByUserName(testUser)
	if user.Credits != previousCreditAmount+10 {
		t.Fatal("failed to refund credits")
	}
	recorder := httptest.NewRecorder()
	testCtx, _ := gin.CreateTestContext(recorder)
	urlValues := url.Values{}
	urlValues.Add("suchkey", "muchvalue")
	testCtx.Request = &http.Request{PostForm: urlValues}
	forms := api.extractPostForms(testCtx, "suchkey")
	if len(forms) == 0 {
		t.Fatal("failed to extract post forms")
	}
	if forms["suchkey"] != "muchvalue" {
		t.Fatal("failed to extract proper postform")
	}
}

func loadDatabase(cfg *config.TemporalConfig) (*gorm.DB, error) {
	return database.OpenDBConnection(database.DBOptions{
		User:           cfg.Database.Username,
		Password:       cfg.Database.Password,
		Address:        cfg.Database.URL,
		Port:           cfg.Database.Port,
		SSLModeDisable: true,
	})
}
