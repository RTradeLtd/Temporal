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
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfscluster"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/gorm"
	"github.com/RTradeLtd/rtfs"
	"github.com/c2h5oh/datasize"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
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

func setupAPI(fakeLens *mocks.FakeLensV2Client, fakeOrch *mocks.FakeServiceClient, fakeSigner *mocks.FakeSignerClient, cfg *config.TemporalConfig, db *gorm.DB) (*API, *httptest.ResponseRecorder, error) {
	dev = true
	// setup connection to ipfs-node-1
	im, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		"", time.Minute*60,
	)
	if err != nil {
		return nil, nil, err
	}
	imCluster, err := rtfscluster.Initialize(
		context.Background(),
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
	clients := Clients{
		Lens:   fakeLens,
		Orch:   fakeOrch,
		Signer: fakeSigner,
	}
	api, err := new(cfg, engine, logger, clients, im, imCluster, false)
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
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}

	api, testRecorder, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		method   string
		call     string
		username string
		password string
		email    string
	}

	// register calls
	tests := []struct {
		name     string
		args     args
		wantCode int
	}{
		{"Register-testuser2", args{"POST", "/v2/auth/register", "testuser2", "password123!@#$%^&&**(!@#!", "testuser@example.org"}, 200},
		{"Register-Email-Fail", args{"POST", "/v2/auth/register", "testuer3", "password123", "testuser+test22@example.org"}, 400},
		{"Register-DuplicateUser", args{"POST", "/v2/auth/register", "testuser2", "password123", "testuser+user22example.org"}, 400},
		{"Register-DuplicateEmail", args{"POST", "/v2/auth/register", "testuser333", "password123", "testuser@example.org"}, 400},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRecorder = httptest.NewRecorder()
			// register user account
			// /v2/auth/register
			var interfaceAPIResp interfaceAPIResponse
			urlValues := url.Values{}
			urlValues.Add("username", tt.args.username)
			urlValues.Add("password", tt.args.password)
			urlValues.Add("email_address", tt.args.email)
			if err := sendRequest(
				api, "POST", "/v2/auth/register", tt.wantCode, nil, urlValues, &interfaceAPIResp,
			); err != nil {
				t.Fatal(err)
			}
			// validate the response code
			if interfaceAPIResp.Code != tt.wantCode {
				t.Fatalf("bad api status code from %s", tt.args.call)
			}
		})
	}

	// login calls
	tests = []struct {
		name     string
		args     args
		wantCode int
	}{
		{"Login-testuser2", args{"POST", "/v2/auth/login", "testuser2", "password123!@#$%^&&**(!@#!", ""}, 200},
		{"Login-testuser", args{"POST", "/v2/auth/login", "testuser", "admin", ""}, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testRecorder = httptest.NewRecorder()
			req := httptest.NewRequest(
				tt.args.method,
				tt.args.call,
				strings.NewReader(fmt.Sprintf("{\n  \"username\": \"%s\",\n  \"password\": \"%s\"\n}", tt.args.username, tt.args.password)),
			)
			api.r.ServeHTTP(testRecorder, req)
			if testRecorder.Code != tt.wantCode {
				t.Fatalf("bad http status code from %s", tt.args.call)
			}
			bodBytes, err := ioutil.ReadAll(testRecorder.Result().Body)
			if err != nil {
				t.Fatal(err)
			}
			// if we're logging in with the account used for testing the v2 api, update our authorization header
			if tt.args.username == "testuser" {
				var loginResp loginResponse
				if err = json.Unmarshal(bodBytes, &loginResp); err != nil {
					t.Fatal(err)
				}
				authHeader = "Bearer " + loginResp.Token
			}
		})
	}
	if str := api.GetIPFSEndpoint("networkName"); str == "" {
		t.Fatal("failed to construct api endpoint")
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
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}

	api, _, err := setupAPI(fakeLens, fakeOrch, fakeSigner, cfg, db)
	if err != nil {
		t.Fatal(err)
	}

	// test systems check route
	// //v2/systems/check
	var apiResp apiResponse
	if err := sendRequest(
		api, "GET", "/v2/systems/check", 200, nil, nil, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// validate the response code
	if apiResp.Code != 200 {
		t.Fatal("bad api status code from /v2/systems/check")
	}
	if apiResp.Response != "systems online" {
		t.Fatal("bad system status recovered")
	}

	// test systems statistics
	// /v2/statistics/stats
	if err := sendRequest(
		api, "GET", "/v2/statistics/stats", 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}

	// test token refresh
	// /v2/auth/refresh
	if err := sendRequest(
		api, "GET", "/v2/auth/refresh", 200, nil, nil, nil,
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
	clients := Clients{
		Lens:   &mocks.FakeLensV2Client{},
		Orch:   &mocks.FakeServiceClient{},
		Signer: &mocks.FakeSignerClient{},
	}
	api, err := Initialize(context.Background(), cfg, "", Options{DevMode: true, DebugLogging: true}, clients, logger)
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
		{"RTCFail", args{"RTC", "bitcoinz"}, true},
		{"RTCPass", args{"rtc", "ethereum"}, false},
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
	forms, missingField := api.extractPostForms(testCtx, "suchkey")
	if missingField != "" {
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
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	clients := Clients{
		Lens:   fakeLens,
		Orch:   fakeOrch,
		Signer: fakeSigner,
	}
	if _, err := Initialize(context.Background(), cfg, "", Options{DevMode: true, DebugLogging: true}, clients, logger); err == nil {
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
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	clients := Clients{
		Lens:   fakeLens,
		Orch:   fakeOrch,
		Signer: fakeSigner,
	}
	if _, err := Initialize(context.Background(), cfg, "", Options{DevMode: true, DebugLogging: true}, clients, logger); err == nil {
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
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	clients := Clients{
		Lens:   fakeLens,
		Orch:   fakeOrch,
		Signer: fakeSigner,
	}
	if _, err := Initialize(context.Background(), cfg, "", Options{DevMode: true, DebugLogging: true}, clients, logger); err == nil {
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
	cfg.Services.Krab.TLS.CertPath = "/root"

	// setup fake mock clients
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	clients := Clients{
		Lens:   fakeLens,
		Orch:   fakeOrch,
		Signer: fakeSigner,
	}
	if _, err := Initialize(context.Background(), cfg, "", Options{DevMode: true, DebugLogging: true}, clients, logger); err == nil {
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
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	clients := Clients{
		Lens:   fakeLens,
		Orch:   fakeOrch,
		Signer: fakeSigner,
	}
	if _, err := Initialize(context.Background(), cfg, "", Options{DevMode: true, DebugLogging: true}, clients, logger); err == nil {
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
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	clients := Clients{
		Lens:   fakeLens,
		Orch:   fakeOrch,
		Signer: fakeSigner,
	}
	api, err := Initialize(context.Background(), cfg, "", Options{DevMode: true, DebugLogging: true}, clients, logger)
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
			fakeLens := &mocks.FakeLensV2Client{}
			fakeOrch := &mocks.FakeServiceClient{}
			fakeSigner := &mocks.FakeSignerClient{}
			clients := Clients{
				Lens:   fakeLens,
				Orch:   fakeOrch,
				Signer: fakeSigner,
			}
			api, err := Initialize(context.Background(), cfg, "", Options{DevMode: true, DebugLogging: true}, clients, logger)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			defer api.Close()
			if tt.args.certFilePath != "" {
				err = api.ListenAndServe(ctx, "127.0.0.1:6700", &TLSConfig{tt.args.certFilePath, tt.args.keyFilePath})
			} else {
				err = api.ListenAndServe(ctx, "127.0.0.1:6701", nil)
			}
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListenAndServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAPI_HandleQueuError_Success(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	logger, err := log.NewLogger("stdout", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup fake mock clients
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	clients := Clients{
		Lens:   fakeLens,
		Orch:   fakeOrch,
		Signer: fakeSigner,
	}
	api, err := Initialize(context.Background(), cfg, "", Options{DevMode: true, DebugLogging: true}, clients, logger)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		if err := api.ListenAndServe(context.Background(), "127.0.0.1:6799", &TLSConfig{"../../testenv/certs/api.cert", "../../testenv/certs/api.key"}); err != nil && err != http.ErrServerClosed {
			t.Fatal(err)
		}
	}()
	type args struct {
		queueType queue.Queue
	}
	tests := []struct {
		name string
		args args
	}{
		{queue.IpfsClusterPinQueue.String(), args{queue.IpfsClusterPinQueue}},
		{queue.EmailSendQueue.String(), args{queue.EmailSendQueue}},
		{queue.IpnsEntryQueue.String(), args{queue.IpnsEntryQueue}},
		{queue.IpfsPinQueue.String(), args{queue.IpfsPinQueue}},
		{queue.IpfsKeyCreationQueue.String(), args{queue.IpfsKeyCreationQueue}},
		{queue.DashPaymentConfirmationQueue.String(), args{queue.DashPaymentConfirmationQueue}},
		{queue.EthPaymentConfirmationQueue.String(), args{queue.EthPaymentConfirmationQueue}},
	}
	// declare an error to use for testing
	amqpErr := &amqp.Error{Code: 400, Reason: "test", Server: true, Recover: false}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test ListenAndServe queue function handling
			switch tt.args.queueType {
			case queue.IpfsClusterPinQueue:
				api.queues.cluster.ErrCh <- amqpErr
			case queue.EmailSendQueue:
				api.queues.email.ErrCh <- amqpErr
			case queue.IpnsEntryQueue:
				api.queues.ipns.ErrCh <- amqpErr
			case queue.IpfsPinQueue:
				api.queues.pin.ErrCh <- amqpErr
			case queue.IpfsKeyCreationQueue:
				api.queues.key.ErrCh <- amqpErr
			case queue.DashPaymentConfirmationQueue:
				api.queues.dash.ErrCh <- amqpErr
			case queue.EthPaymentConfirmationQueue:
				api.queues.eth.ErrCh <- amqpErr
			}
			// test handleQueueError function directly
			if _, err := api.handleQueueError(amqpErr, api.cfg.RabbitMQ.URL, tt.args.queueType, true); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAPI_HandleQueuError_Failure(t *testing.T) {
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	logger, err := log.NewLogger("stdout", true)
	if err != nil {
		t.Fatal(err)
	}
	// setup fake mock clients
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	clients := Clients{
		Lens:   fakeLens,
		Orch:   fakeOrch,
		Signer: fakeSigner,
	}
	api, err := Initialize(context.Background(), cfg, "", Options{DevMode: true, DebugLogging: true}, clients, logger)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		queueType queue.Queue
	}
	tests := []struct {
		name string
		args args
	}{
		{queue.IpfsClusterPinQueue.String(), args{queue.IpfsClusterPinQueue}},
		{queue.EmailSendQueue.String(), args{queue.EmailSendQueue}},
		{queue.IpnsEntryQueue.String(), args{queue.IpnsEntryQueue}},
		{queue.IpfsPinQueue.String(), args{queue.IpfsPinQueue}},
		{queue.IpfsKeyCreationQueue.String(), args{queue.IpfsKeyCreationQueue}},
		{queue.DashPaymentConfirmationQueue.String(), args{queue.DashPaymentConfirmationQueue}},
		{queue.EthPaymentConfirmationQueue.String(), args{queue.EthPaymentConfirmationQueue}},
	}
	// setup a bad rabbitmq url for testing connectivity failures
	api.cfg.RabbitMQ.URL = "notarealprotocol://notarealurl"
	// declare an error to use for testing
	amqpErr := &amqp.Error{Code: 400, Reason: "test", Server: true, Recover: false}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//TODO: enable this kind of testing within ListenAndServe eventually
			// test handleQueueError function directly
			if _, err := api.handleQueueError(amqpErr, api.cfg.RabbitMQ.URL, tt.args.queueType, true); err == nil {
				t.Fatal("error expected")
			}
		})
	}
}

func loadDatabase(cfg *config.TemporalConfig) (*gorm.DB, error) {
	dbm, err := database.New(cfg, database.Options{SSLModeDisable: true})
	if err != nil {
		return nil, err
	}
	return dbm.DB, nil
}
