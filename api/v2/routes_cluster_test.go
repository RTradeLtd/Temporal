package v2

import (
	"net/url"
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config"
)

func Test_API_Routes_Cluster(t *testing.T) {
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

	// test cluster pin
	// /api/v2/ipfs/cluster/pin
	var apiResp apiResponse
	urlValues := url.Values{}
	urlValues.Add("hold_time", "5")
	if err := sendRequest(
		api, "POST", "/api/v2/ipfs/cluster/pin/"+hash, 200, nil, urlValues, &apiResp,
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
}
