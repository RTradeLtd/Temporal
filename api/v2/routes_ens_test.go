package v2

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config/v2"
	"github.com/gin-gonic/gin"
)

func Test_API_Routes_ENS(t *testing.T) {
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
	fakeWalletService := &mocks.FakeWalletServiceClient{}
	// instantiate the test api
	api, testRecorder, err := setupAPI(t, fakeLens, fakeOrch, fakeSigner, fakeWalletService, cfg, db)
	if err != nil {
		t.Fatal(err)
	}
	testRecorder = httptest.NewRecorder()
	gctx, _ := gin.CreateTestContext(testRecorder)
	req := httptest.NewRequest("GET", "/v2/ens/claim", nil)
	vals := url.Values{}
	req.PostForm = vals
	req.Header.Add("Authorization", authHeader)
	gctx.Request = req
	api.ClaimENSName(gctx)

	testRecorder = httptest.NewRecorder()
	gctx, _ = gin.CreateTestContext(testRecorder)
	req = httptest.NewRequest("GET", "/v2/ens/update", nil)
	vals = url.Values{}
	vals.Add("content_hash", hash)
	req.PostForm = vals
	req.Header.Add("Authorization", authHeader)
	gctx.Request = req
	api.usage.ClaimENSName("testuser")
	api.UpdateContentHash(gctx)
}
