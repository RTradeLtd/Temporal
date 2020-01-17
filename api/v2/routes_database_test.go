package v2

import (
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/config/v2"
)

func Test_API_Routes_Database(t *testing.T) {
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

	api, err := setupAPI(t, fakeLens, fakeOrch, fakeSigner, fakeWalletService, cfg, db)
	if err != nil {
		t.Fatal(err)
	}

	// test database specific uploads
	// /v2/database/uploads/testuser
	var interfaceAPIResp interfaceAPIResponse
	if err := sendRequest(
		api, "GET", "/v2/database/uploads", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}
	// test paginated
	interfaceAPIResp = interfaceAPIResponse{}
	if err := sendRequest(
		api, "GET", "/v2/database/uploads?paged=true", 200, nil, nil, &interfaceAPIResp,
	); err != nil {
		t.Fatal(err)
	}

	// test get encrypted uploads
	// /v2/frontend/uploads/encrypted
	if err := sendRequest(
		api, "GET", "/v2/database/uploads/encrypted", 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}
	if err := sendRequest(
		api, "GET", "/v2/database/uploads/encrypted?paged=true", 200, nil, nil, nil,
	); err != nil {
		t.Fatal(err)
	}
}
