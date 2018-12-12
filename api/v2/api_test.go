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
	"github.com/RTradeLtd/Temporal/rtfscluster"
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
	hash         = "QmPY5iMFjNZKxRbUZZC85wXb9CFgNSyzAy1LxwL62D8VGr"
	api          *API
	db           *gorm.DB
	engine       *gin.Engine
	testRecorder *httptest.ResponseRecorder
	authHeader   string
	cfg          *config.TemporalConfig
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

type boolAPIResponse struct {
	Code     int  `json:"code"`
	Response bool `json:"bool"`
}

type mapAPIResponse struct {
	Code     int   `json:"code"`
	Response gin.H `json:"response"`
}

type interfaceAPIResponse struct {
	Code     int         `json:"code"`
	Response interface{} `json:"response"`
}

type floatAPIResponse struct {
	Code     int     `json:"code"`
	Response float64 `json:"response"`
}

type ipnsAPIResponse struct {
	Code     int            `json:"code"`
	Response *[]models.IPNS `json:"response"`
}

type stringSliceAPIResponse struct {
	Code     int      `json:"code"`
	Response []string `json:"response"`
}

// sendRequest is a helper method used to handle sending an api request
func sendRequest(method, url string, wantStatus int, body io.Reader, urlValues url.Values, out interface{}) error {
	testRecorder := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, body)
	req.Header.Add("Authorization", authHeader)
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != wantStatus {
		return fmt.Errorf("received status %v expected %v from api call %s", testRecorder.Code, wantStatus, url)
	}
	if out == nil {
		return nil
	}
	// unmarshal the response
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bodyBytes, out)
}

func Test_API_Setup(t *testing.T) {
	var err error
	// load configuration
	cfg, err = config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err = loadDatabase(cfg)
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
	imCluster, err := rtfscluster.Initialize(
		cfg.IPFSCluster.APIConnection.Host,
		cfg.IPFSCluster.APIConnection.Port,
	)
	if err != nil {
		t.Fatal(err)
	}
	// create our test api
	testRecorder = httptest.NewRecorder()
	_, engine = gin.CreateTestContext(testRecorder)
	api, err = new(cfg, engine, &mocks.FakeIndexerAPIClient{}, im, imCluster, false, os.Stdout)
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
	authHeader = "Bearer " + loginResp.Token
}

func Test_API_Routes_Lens(t *testing.T) {
	// test lens index - valid object type
	// /api/v2/lens/index
	urlValues := url.Values{}
	urlValues.Add("object_type", "ipld")
	urlValues.Add("object_identifier", hash)
	if err := sendRequest(
		"POST", "/api/v2/lens/index", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test lens index - invalid object type
	// /api/v2/lens/index
	urlValues = url.Values{}
	urlValues.Add("object_type", "storj")
	urlValues.Add("object_identifier", hash)
	if err := sendRequest(
		"POST", "/api/v2/lens/index", 400, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test lens index - bad format hash
	// /api/v2/lens/index
	urlValues = url.Values{}
	urlValues.Add("object_type", "ipld")
	urlValues.Add("object_identifier", "notarealipfshash")
	if err := sendRequest(
		"POST", "/api/v2/lens/index", 400, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test lens search
	// /api/v2/lens/search
	urlValues = url.Values{}
	urlValues.Add("keywords", "minivan")
	urlValues.Add("keywords", "protocols")
	if err := sendRequest(
		"POST", "/api/v2/lens/search", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}
}

func Test_API_Routes_Payments(t *testing.T) {
	// test basic dash payment
	// /api/v2/payments/create/dash
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v2/payments/create/dash", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues := url.Values{}
	urlValues.Add("credit_value", "10")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test request signed payment message - rtc
	// /api/v2/payments/request
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/payments/request", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("payment_type", "rtc")
	urlValues.Add("sender_address", "0x0")
	urlValues.Add("credit_value", "10")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test request signed payment message - eth
	// /api/v2/payments/request
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/payments/request", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("payment_type", "eth")
	urlValues.Add("sender_address", "0x0")
	urlValues.Add("credit_value", "10")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test payment confirmation
	// /api/v2/payments/confirm
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/payments/confirm", nil)
	req.Header.Add("Authorization", authHeader)
	urlValues = url.Values{}
	urlValues.Add("payment_number", "10")
	urlValues.Add("tx_hash", "0x1")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)

	// test valid deposit address
	args := []string{"eth", "rtc", "btc", "ltc", "xmr", "dash"}
	for _, v := range args {
		if err := sendRequest(
			"GET", "/api/v2/payments/deposit/address/"+v, 200, nil, nil, nil,
		); err != nil {
			t.Fatal(err)
		}
	}

	// test invalid deposit address
	if err := sendRequest(
		"GET", "/api/v2/payments/deposit/address/invalidType", 400, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}
}

func Test_API_Routes_Misc(t *testing.T) {
	// test systems check route
	// //api/v2/systems/check
	var apiResp apiResponse
	if err := sendRequest(
		"GET", "/api/v2/systems/check", 200, nil, nil, &apiResp,
	); err != nil {
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

	if err := sendRequest(
		"GET", "/api/v2/statistics/stats", 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test mini create bucket
	// /api/v2/admin/mini/create/bucket
	if err := os.Setenv("MINI_SSL_ENABLE", "false"); err != nil {
		t.Fatal(err)
	}
	urlValues := url.Values{}
	urlValues.Add("bucket_name", "filesuploadbucket")
	if err := sendRequest(
		"POST", "/api/v2/admin/mini/create/bucket", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}
}

func Test_API_Routes_IPFS_Public(t *testing.T) {
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
	req := httptest.NewRequest("POST", "/api/v2/ipfs/public/file/add", bodyBuf)
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
	bodyBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
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
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/public/pin/"+hash, 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/ipfs/public/pin")
	}

	// test pin check
	// /api/v2/ipfs/public/check/pin
	var boolAPIResp boolAPIResponse
	if err := sendRequest(
		"GET", "/api/v2/ipfs/public/pin/check/"+hash, 200, nil, nil, &boolAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if boolAPIResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/ipfs/public/check/pin")
	}

	// test pubsub publish
	// /api/v2/ipfs/pubsub/publish/topic
	urlValues = url.Values{}
	urlValues.Add("message", "bar")
	var mapAPIResp mapAPIResponse
	if err := sendRequest(
		"POST", "/api/v2/ipfs/public/pubsub/publish/foo", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/pubsub/publish/topic")
	}
	if mapAPIResp.Response["topic"] != "foo" {
		t.Fatal("bad response")
	}
	if mapAPIResp.Response["message"] != "bar" {
		t.Fatal("bad response")
	}

	// test object stat
	// /api/v2/ipfs/stat
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		"GET", "/api/v2/ipfs/public/stat/"+hash, 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/public/stat")
	}

	// test get dag
	// /api/v2/ipfs/public/dag
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		"GET", "/api/v2/ipfs/public/dag/"+hash, 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/public/dag/")
	}

	// test download
	// /api/v2/ipfs/public/download
	if err := sendRequest(
		"POST", "/api/v2/ipfs/public/download/"+hash, 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test public network beam
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "public")
	urlValues.Add("destination_network", "public")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}
}

func Test_API_Routes_IPFS_Private(t *testing.T) {
	// create a private network for us to test against
	nm := models.NewHostedIPFSNetworkManager(db)
	um := models.NewUserManager(db)
	if _, err := nm.CreateHostedPrivateNetwork("abc123", "abc123", nil, []string{"testuser"}); err != nil {
		t.Fatal(err)
	}
	if err := um.AddIPFSNetworkForUser("testuser", "abc123"); err != nil {
		t.Fatal(err)
	}
	if err := nm.UpdateNetworkByName("abc123", map[string]interface{}{
		"api_url": cfg.IPFS.APIConnection.Host + ":" + cfg.IPFS.APIConnection.Port,
	}); err != nil {
		t.Fatal(err)
	}

	// get private network information
	// /api/v2/ipfs/private/network
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		"GET", "/api/v2/ipfs/private/network/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /api/v2/ipfs/private/network/abc123")
	}

	// get all authorized private networks
	// /api/v2/ipfs/private/networks
	var stringSliceAPIResp stringSliceAPIResponse
	if err := sendRequest(
		"GET", "/api/v2/ipfs/private/networks", 200, nil, nil, &stringSliceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if stringSliceAPIResp.Code != 200 {
		t.Fatal("bad api response status code from /api/v2/ipfs/private/networks")
	}
	if len(stringSliceAPIResp.Response) == 0 {
		t.Fatal("failed to find any from /api/v2/ipfs/private/networks")
	}
	var found bool
	for _, v := range stringSliceAPIResp.Response {
		if v == "abc123" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("failed to find correct network from /api/v2/ipfs/private/networks")
	}

	// stop private network
	// /api/v2/ipfs/private/network/stop
	// for now until we implement proper grpc testing, this will fail
	var mapAPIResp mapAPIResponse
	urlValues := url.Values{}
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/private/network/stop", 400, nil, urlValues, &mapAPIResp,
	); err == nil {
		t.Fatal("failed to properly handle api call")
	}

	// start private network
	// /api/v2/ipfs/private/network/start
	// for now until we implement proper grpc testing, this will fail
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/private/network/start", 400, nil, urlValues, &mapAPIResp,
	); err == nil {
		t.Fatal("failed to properly handle api call")
	}

	// add a file normally
	// /api/v2/ipfs/private/file/add
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
	req := httptest.NewRequest("POST", "/api/v2/ipfs/private/file/add", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	urlValues.Add("network_name", "abc123")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /api/v2/ipfs/private/file/add")
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
		t.Fatal("bad api status code from /api/v2/ipfs/private/file/add")
	}
	hash = apiResp.Response

	// add a file advanced
	// /api/v2/ipfs/private/file/add/advanced
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
	req = httptest.NewRequest("POST", "/api/v2/ipfs/private/file/add/advanced", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	urlValues.Add("passphrase", "password123")
	urlValues.Add("network_name", "abc123")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /api/v2/ipfs/private/file/add/advanced")
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
		t.Fatal("bad api status code from /api/v2/ipfs/private/file/add/advanced")
	}

	// test pinning
	// /api/v2/ipfs/private/pin
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("hold_time", "5")
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/private/pin/"+hash, 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/ipfs/private/pin")
	}

	// test pin check
	// /api/v2/ipfs/private/check/pin
	var boolAPIResp boolAPIResponse
	if err := sendRequest(
		"GET", "/api/v2/ipfs/private/pin/check/"+hash+"/abc123", 200, nil, nil, &boolAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if boolAPIResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/ipfs/private/check/pin")
	}

	// test pubsub publish
	// /api/v2/ipfs/private/publish/topic
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("message", "bar")
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/private/pubsub/publish/foo", 200, nil, urlValues, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api status code from  /api/v2/ipfs/private/pubsub/publish/topic")
	}
	if mapAPIResp.Response["topic"] != "foo" {
		t.Fatal("bad response")
	}
	if mapAPIResp.Response["message"] != "bar" {
		t.Fatal("bad response")
	}

	// test object stat
	// /api/v2/ipfs/private/stat
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		"GET", "/api/v2/ipfs/private/stat/"+hash+"/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/private/stat")
	}

	// test get dag
	// /api/v2/ipfs/private/dag
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		"GET", "/api/v2/ipfs/private/dag/"+hash+"/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/private/dag/")
	}

	// test download
	// /api/v2/ipfs/public/download
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/private/download/"+hash, 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test get authorized networks
	// /api/v2/ipfs/private/networks
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		"GET", "/api/v2/ipfs/private/networks", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/private/networks/")
	}

	// test get authorized networks
	// /api/v2/ipfs/private/networks
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		"GET", "/api/v2/ipfs/private/uploads/abc123", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad response status code from /api/v2/ipfs/private/uploads")
	}

	// test private network beam - source private, dest public
	urlValues = url.Values{}
	urlValues.Add("source_network", "abc123")
	urlValues.Add("destination_network", "public")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test private network beam - source public, dest private
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "public")
	urlValues.Add("destination_network", "abc123")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test private network beam - source private, dest private
	// /api/v2/ipfs/utils/laser/beam
	urlValues = url.Values{}
	urlValues.Add("source_network", "abc123")
	urlValues.Add("destination_network", "abc123")
	urlValues.Add("content_hash", hash)
	urlValues.Add("passphrase", "password123")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/utils/laser/beam", 200, nil, urlValues, nil,
	); err != nil {
		t.Fatal(err)
	}

	// remove private network
	// /api/v2/ipfs/private/network/remove
	// for now until we implement proper grpc testing, this will fail
	mapAPIResp = mapAPIResponse{}
	urlValues = url.Values{}
	urlValues.Add("network_name", "abc123")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/private/network/remove", 400, nil, urlValues, &mapAPIResp,
	); err == nil {
		t.Fatal("failed to properly handle api call")
	}
}

func Test_API_Routes_IPNS(t *testing.T) {
	// test get ipns records
	// /api/v2/ipns/records
	testRecorder = httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v2/ipns/records", nil)
	req.Header.Add("Authorization", authHeader)
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/ipns/records")
	}

	// test ipns publishing (public)
	// api/v2/ipns/public/publish/details
	// generate a fake key to publish
	var apiResp apiResponse
	um := models.NewUserManager(db)
	um.AddIPFSKeyForUser("testuser", "mytestkey", "suchkeymuchwow")
	urlValues := url.Values{}
	urlValues.Add("hash", hash)
	urlValues.Add("life_time", "24h")
	urlValues.Add("ttl", "1h")
	urlValues.Add("key", "mytestkey")
	urlValues.Add("resolve", "true")
	if err := sendRequest(
		"POST", "/api/v2/ipns/public/publish/details", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipns/public/publish/details")
	}

	// test ipns publishing (private)
	// api/v2/ipns/private/publish/details
	// create a fake private network
	apiResp = apiResponse{}
	nm := models.NewHostedIPFSNetworkManager(db)
	if _, err := nm.CreateHostedPrivateNetwork("testnetwork", "fakeswarmkey", nil, []string{"testuser"}); err != nil {
		t.Fatal(err)
	}
	if err := um.AddIPFSNetworkForUser("testuser", "testnetwork"); err != nil {
		t.Fatal(err)
	}
	urlValues = url.Values{}
	urlValues.Add("hash", hash)
	urlValues.Add("life_time", "24h")
	urlValues.Add("ttl", "1h")
	urlValues.Add("key", "mytestkey")
	urlValues.Add("resolve", "true")
	urlValues.Add("network_name", "testnetwork")
	if err := sendRequest(
		"POST", "/api/v2/ipns/private/publish/details", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipns/private/publish/details")
	}

	// test get ipns records
	// /api/v2/ipns/records
	// spoof a fake record as we arent using the queues in this test
	var ipnsAPIResp ipnsAPIResponse
	im := models.NewIPNSManager(db)
	if _, err := im.CreateEntry("fakekey", "fakehash", "fakekeyname", "public", "testuser", time.Minute, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := sendRequest(
		"GET", "/api/v2/ipns/records", 200, nil, nil, &ipnsAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if ipnsAPIResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipns/private/publish/details")
	}
	if len(*ipnsAPIResp.Response) == 0 {
		t.Fatal("no records discovered")
	}
}

func Test_API_Routes_Cluster(t *testing.T) {
	// test cluster sync
	// /api/v2/ipfs/cluster/sync/errors/local
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		"POST", "/api/v2/ipfs/cluster/sync/errors/local", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipfs/cluster/sync/errors/local")
	}

	// test cluster pin
	// /api/v2/ipfs/cluster/pin
	var apiResp apiResponse
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	if err := sendRequest(
		"POST", "/api/v2/ipfs/cluster/pin/"+hash, 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipfs/cluster/pin")
	}
	// manually pin since we aren't using queues
	decoded, err := api.ipfsCluster.DecodeHashString(hash)
	if err != nil {
		t.Fatal(err)
	}
	if err = api.ipfsCluster.Pin(decoded); err != nil {
		t.Fatal(err)
	}

	// test cluster local status
	// /api/v2/ipfs/cluster/status/pin/local
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		"GET", "/api/v2/ipfs/cluster/status/pin/local/"+hash, 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipfs/cluster/status/pin/local")
	}

	// test cluster local status
	// /api/v2/ipfs/cluster/status/pin/global
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		"GET", "/api/v2/ipfs/cluster/status/pin/global/"+hash, 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipfs/cluster/status/pin/global")
	}

	// test cluster local status
	// /api/v2/ipfs/cluster/status/local
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		"GET", "/api/v2/ipfs/cluster/status/local", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/ipfs/cluster/status/local")
	}
}

func Test_API_Routes_Frontend(t *testing.T) {
	// test get encrypted uploads
	// /api/v2/frontend/uploads/encrypted
	if err := sendRequest(
		"GET", "/api/v2/frontend/uploads/encrypted", 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test pin cost calculate
	// /api/v2/frontend/cost/calculate/:hash/:holTime
	var floatAPIResp floatAPIResponse
	if err := sendRequest(
		"GET", "/api/v2/frontend/cost/calculate/"+hash+"/5", 200, nil, nil, &floatAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	if floatAPIResp.Code != 200 {
		t.Fatal("bad response code from /api/v2/frontend/cost/calculate")
	}
	if floatAPIResp.Response == 0 {
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
	req := httptest.NewRequest("POST", "/api/v2/frontend/cost/calculate/file", bodyBuf)
	req.Header.Add("Authorization", authHeader)
	req.Header.Add("Content-Type", bodyWriter.FormDataContentType())
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	req.PostForm = urlValues
	api.r.ServeHTTP(testRecorder, req)
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code recovered from /api/v2/frontend/cost/calculate/file")
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
		t.Fatal("bad api status code from /api/v2/frontend/cost/calculate/file")
	}
	if floatAPIResp.Response == 0 {
		t.Fatal("failed to calculate cost /api/v2/frontend/cost/calculate/file")
	}
}

func Test_API_Routes_Database(t *testing.T) {
	// test database global uploads
	// /api/v2/database/uploads
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		"GET", "/api/v2/database/uploads", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api status code from api/v2/database/uploads")
	}

	// test database specific uploads
	// /api/v2/database/uploads/testuser
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		"GET", "/api/v2/database/uploads/testuser", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api status code from api/v2/database/uploads")
	}
}

func Test_API_Routes_Account(t *testing.T) {
	// verify the username from the token
	// /api/v2/account/token/username
	var apiResp apiResponse
	if err := sendRequest(
		"GET", "/api/v2/account/token/username", 200, nil, nil, &apiResp,
	); err != nil {
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
	apiResp = apiResponse{}
	if err := sendRequest(
		"GET", "/api/v2/account/email/token/get", 200, nil, nil, &apiResp,
	); err != nil {
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
	urlValues := url.Values{}
	urlValues.Add("token", user.EmailVerificationToken)
	apiResp = apiResponse{}
	if err := sendRequest(
		"POST", "/api/v2/account/email/token/verify", 200, nil, urlValues, &apiResp,
	); err != nil {
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
	urlValues = url.Values{}
	urlValues.Add("old_password", "admin")
	urlValues.Add("new_password", "admin1234@")
	apiResp = apiResponse{}
	if err := sendRequest(
		"POST", "/api/v2/account/password/change", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/password/change")
	}

	// create ipfs keys
	// /api/v2/account/key/ipfs/new
	urlValues = url.Values{}
	urlValues.Add("key_type", "ed25519")
	urlValues.Add("key_bits", "256")
	urlValues.Add("key_name", "key1")
	apiResp = apiResponse{}
	if err := sendRequest(
		"POST", "/api/v2/account/key/ipfs/new", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/new")
	}
	// test rsa keys
	urlValues.Add("key_type", "rsa")
	urlValues.Add("key_bits", "2048")
	urlValues.Add("key_name", "key2")
	apiResp = apiResponse{}
	if err := sendRequest(
		"POST", "/api/v2/account/key/ipfs/new", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/new")
	}
	if apiResp.Code != 200 {
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
	var mapAPIResp mapAPIResponse
	if err := sendRequest(
		"GET", "/api/v2/account/key/ipfs/get", 200, nil, nil, &mapAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if mapAPIResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/key/ipfs/get")
	}

	// get available credits
	// /api/v2/account/credits/available
	var floatAPIResp floatAPIResponse
	if err := sendRequest(
		"GET", "/api/v2/account/credits/available", 200, nil, nil, &floatAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if floatAPIResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/credits/available")
	}

	// register user account
	// /api/v2/auth/register
	var interfaceAPIResp interfaceAPIResponse
	urlValues = url.Values{}
	urlValues.Add("username", "testuser2")
	urlValues.Add("password", "password123")
	urlValues.Add("email_address", "testuser2+test@example.org")
	if err := sendRequest(
		"POST", "/api/v2/auth/register", 200, nil, urlValues, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/auth/register")
	}

	// forgot email
	// /api/v2/account/email/forgot
	apiResp = apiResponse{}
	if err := sendRequest(
		"POST", "/api/v2/account/email/forgot", 200, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/account/email/forgot")
	}

	// test@email.com
	// forgot username
	// /api/v2/forgot/username
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("email_address", "test@email.com")
	if err := sendRequest(
		"POST", "/api/v2/forgot/username", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/forgot/username")
	}

	// test@email.com
	// forgot password
	// /api/v2/forgot/password
	apiResp = apiResponse{}
	urlValues = url.Values{}
	urlValues.Add("email_address", "test@email.com")
	if err := sendRequest(
		"POST", "/api/v2/forgot/password", 200, nil, urlValues, &apiResp,
	); err != nil {
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
	api, err := Initialize(cfg, true, &mocks.FakeIndexerAPIClient{})
	if err != nil {
		t.Fatal(err)
	}
	if err := api.FileSizeCheck(int64(datasize.GB.Bytes() * 1)); err != nil {
		t.Fatal(err)
	}
	if err := api.FileSizeCheck(int64(datasize.GB.Bytes() * 1)); err != nil {
		t.Fatal(err)
	}
	if err := api.FileSizeCheck(int64(datasize.GB.Bytes() * 10)); err == nil {
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
			if _, err := api.getUSDValue(tt.args.paymentType); (err != nil) != tt.wantErr {
				t.Errorf("getUSDValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	if err := api.validateUserCredits(testUser, 1); err != nil {
		t.Fatal(err)
	}
	if err := api.validateUserCredits(testUser, tooManyCredits); err == nil {
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