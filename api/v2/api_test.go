package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	log "github.com/RTradeLtd/Temporal/log"
	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/Temporal/rtfscluster"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/rtfs"
	"github.com/c2h5oh/datasize"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

const (
	tooManyCredits     = 10.9999997e+07
	testUser           = "testuser"
	testSwarmKey       = "7fcb5a1b19bdda69da7307162e3becd2d6bd485d5aad778470b305f3f306cf79"
	testBootstrapPeer1 = "/ip4/172.218.49.115/tcp/5002/ipfs/Qmf964tiE9JaxqntDsSBGasD4aaofPQtfYZyMSJJkRrVTQ"
	testBootstrapPeer2 = "/ip4/192.168.1.249/tcp/4001/ipfs/QmXuGVPzEz2Ji7g54AYyqoobRJNHqtnrfaEceAes2bTKMh"
)

var (
	hash       = "QmS4ustL54uo8FzR9455qaxZwuMiUhyvMcX9Ba8nUH4uVv"
	authHeader string
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

type lensSearchAPIResponse struct {
	Code     int                 `json:"code"`
	Response []map[string]string `json:"response"`
}

// sendRequest is a helper method used to handle sending an api request
func sendRequest(api *API, method, url string, wantStatus int, body io.Reader, urlValues url.Values, out interface{}) error {
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

func setupAPI(fakeLens *mocks.FakeIndexerAPIClient, fakeOrch *mocks.FakeServiceClient, fakeSigner *mocks.FakeSignerClient, cfg *config.TemporalConfig, db *gorm.DB) (*API, *httptest.ResponseRecorder, error) {
	dev = true
	// setup connection to ipfs-node-1
	im, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		nil,
		time.Minute*10,
	)
	if err != nil {
		return nil, nil, err
	}
	imCluster, err := rtfscluster.Initialize(
		cfg.IPFSCluster.APIConnection.Host,
		cfg.IPFSCluster.APIConnection.Port,
	)
	if err != nil {
		return nil, nil, err
	}
	// create our test api
	testRecorder := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(testRecorder)
	logger, err := log.NewLogger("stdout", true)
	if err != nil {
		return nil, nil, err
	}

	api, err := new(cfg, engine, logger, fakeLens, fakeOrch, fakeSigner, im, imCluster, false)
	if err != nil {
		return nil, nil, err
	}
	// setup api routes
	if err = api.setupRoutes(); err != nil {
		return nil, nil, err
	}
	return api, testRecorder, nil
}

// this does a quick initial test of the API, and setups a second user account to use for testing
func Test_API_Setup(t *testing.T) {
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

	// register user account
	// /api/v2/auth/register
	var interfaceAPIResp interfaceAPIResponse
	urlValues := url.Values{}
	urlValues.Add("username", "testuser2")
	urlValues.Add("password", "password123!@#$%^&&**(!@#!")
	urlValues.Add("email_address", "testuser2+test@example.org")
	if err := sendRequest(
		api, "POST", "/api/v2/auth/register", 200, nil, urlValues, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if interfaceAPIResp.Code != 200 {
		t.Fatal("bad api status code from /api/v2/auth/register")
	}
	testRecorder = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader("{\n  \"username\": \"testuser2\",\n  \"password\": \"password123!@#$%^&&**(!@#!\"\n}"))
	req.Header.Add("Content-Type", "application/json")
	api.r.ServeHTTP(testRecorder, req)
	// validate the http status code
	if testRecorder.Code != 200 {
		t.Fatal("bad http status code from /api/v2/auth/login")
	}
}

func Test_API_Routes_Misc(t *testing.T) {
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

	// test systems check route
	// //api/v2/systems/check
	var apiResp apiResponse
	if err := sendRequest(
		api, "GET", "/api/v2/systems/check", 200, nil, nil, &apiResp,
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
		api, "GET", "/api/v2/statistics/stats", 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}
}

func Test_Utils(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	logger, err := log.NewLogger("stdout", true)
	if err != nil {
		t.Fatal(err)
	}
	api, err := Initialize(cfg, logger, true, &mocks.FakeIndexerAPIClient{}, &mocks.FakeServiceClient{}, &mocks.FakeSignerClient{})
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
	if err != nil {
		t.Fatal(err)
	}
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

func Test_API_Initialize_Cluster_Failure(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	logger, err := log.NewLogger("stdout", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup an unreachable cluster host
	cfg.IPFSCluster.APIConnection.Host = "10.255.255.255"

	// setup fake mock clients
	fakeLens := &mocks.FakeIndexerAPIClient{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	if _, err := Initialize(cfg, logger, true, fakeLens, fakeOrch, fakeSigner); err == nil {
		t.Fatal("expected error")
	}
}

func Test_API_Initialize_IPFS_Failure(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	logger, err := log.NewLogger("stdout", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup an unreachable cluster host
	cfg.IPFS.APIConnection.Host = "notarealip"

	// setup fake mock clients
	fakeLens := &mocks.FakeIndexerAPIClient{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	if _, err := Initialize(cfg, logger, true, fakeLens, fakeOrch, fakeSigner); err == nil {
		t.Fatal("expected error")
	}
}

func Test_API_Initialize_Setup_Routes_Failure(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	logger, err := log.NewLogger("stdout", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup an invalid connection limit
	cfg.API.Connection.Limit = "notanumber"

	// setup fake mock clients
	fakeLens := &mocks.FakeIndexerAPIClient{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}

	if _, err := Initialize(cfg, logger, true, fakeLens, fakeOrch, fakeSigner); err == nil {
		t.Fatal("expected error")
	}
}

func Test_API_Initialize_Kaas_Failure(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	logger, err := log.NewLogger("stdout", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup an invalid connection limit
	cfg.Endpoints.Krab.TLS.CertPath = "/root"

	// setup fake mock clients
	fakeLens := &mocks.FakeIndexerAPIClient{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}

	if _, err := Initialize(cfg, logger, true, fakeLens, fakeOrch, fakeSigner); err == nil {
		t.Fatal("expected error")
	}
}

func Test_API_Initialize_Queue_Failure(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	logger, err := log.NewLogger("", true)
	if err != nil {
		t.Fatal(err)
	}
	cfg.RabbitMQ.URL = "notarealip"

	// setup fake mock clients
	fakeLens := &mocks.FakeIndexerAPIClient{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}

	if _, err := Initialize(cfg, logger, true, fakeLens, fakeOrch, fakeSigner); err == nil {
		t.Fatal("expected error")
	}
}

func Test_API_Initialize_Main_Network(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	logger, err := log.NewLogger("stdout", true)
	if err != nil {
		t.Fatal(err)
	}
	dev = false

	// setup fake mock clients
	fakeLens := &mocks.FakeIndexerAPIClient{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}

	api, err := Initialize(cfg, logger, true, fakeLens, fakeOrch, fakeSigner)
	if err != nil {
		t.Fatal(err)
	}
	api.Close()
}

func Test_API_Initialize_ListenAndServe(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	logger, err := log.NewLogger("stdout", true)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		certFilePath string
		keyFilePath  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"NoTLS", args{"", ""}, false},
		{"TLS", args{"../../testenv/certs/api.cert", "../../testenv/certs/api.key"}, false},
		{"TLS-Missing-Cert", args{"../../README.md", "../../testenv/certs/api.key"}, true},
		{"TLS-Missing-Key", args{"../../testenv/certs/api.cert", "../../README.md"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeLens := &mocks.FakeIndexerAPIClient{}
			fakeOrch := &mocks.FakeServiceClient{}
			fakeSigner := &mocks.FakeSignerClient{}
			api, err := Initialize(cfg, logger, true, fakeLens, fakeOrch, fakeSigner)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			defer api.Close()
			if tt.args.certFilePath != "" {
				err = api.ListenAndServe(ctx, "127.0.0.1:6700", &TLSFiles{tt.args.certFilePath, tt.args.keyFilePath})
			} else {
				err = api.ListenAndServe(ctx, "127.0.0.1:6701", nil)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListenAndServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
