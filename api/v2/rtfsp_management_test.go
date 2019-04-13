package v2

import (
	"net/url"
	"testing"

	"github.com/RTradeLtd/database/models"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
)

func Test_API_Routes_IPFS_Private_User_Management(t *testing.T) {
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

	if _, err := api.um.NewUserAccount("testaccount2323", "password123", "example@example.org"); err != nil {
		t.Fatal(err)
	}
	if _, err := api.nm.CreateHostedPrivateNetwork(
		"testnetworkdude",
		"swarmkey",
		nil,
		models.NetworkAccessOptions{Owner: "testuser", Users: []string{"testuser"}},
	); err != nil {
		t.Fatal(err)
	}
	var apiResp = apiResponse{}
	urlValues := url.Values{}
	urlValues.Add("network_name", "testnetworkdude")
	urlValues.Add("users", "testaccount2323")
	urlValues.Add("owners", "testaccount2323")
	// test adding a user
	if err := sendRequest(
		api, "POST", "/v2/ipfs/private/network/users/add", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// test removing a user
	if err := sendRequest(
		api, "DELETE", "/v2/ipfs/private/network/users/remove", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
	// test adding an owner
	if err := sendRequest(
		api, "POST", "/v2/ipfs/private/network/owners/add", 200, nil, urlValues, &apiResp,
	); err != nil {
		t.Fatal(err)
	}
}
