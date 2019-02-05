package v2

import (
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database/models"
)

func TestEmailJWT(t *testing.T) {
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
	randUtils := utils.GenerateRandomUtils()
	// create a user for this test
	randEmail := randUtils.GenerateString(32, utils.LetterBytes) + "@example.org"
	randUser := randUtils.GenerateString(32, utils.LetterBytes)
	if _, err := api.um.NewUserAccount(randUser, "password123", randEmail); err != nil {
		t.Fatal(err)
	}
	// generate the email verification string
	userModel, err := api.um.GenerateEmailVerificationToken(randUser)
	if err != nil {
		t.Fatal(err)
	}
	tkn, err := api.generateEmailJWTToken(randUser, userModel.EmailVerificationToken)
	if err != nil {
		t.Fatal(err)
	}
	if err := api.verifyEmailJWTToken(tkn, randUser); err != nil {
		t.Fatal(err)
	}
}

func Test_CheckAccessForPrivateNetwork(t *testing.T) {
	// load configuration
	cfg, err := config.LoadConfig("../../testenv/config.json")
	if err != nil {
		t.Fatal(err)
	}
	db, err := loadDatabase(cfg)
	if err != nil {
		t.Fatal(err)
	}
	um := models.NewUserManager(db)
	if err := um.AddIPFSNetworkForUser("testuser", "mynewnetworktotestwith"); err != nil {
		t.Fatal(err)
	}
	// search for a non-existent user
	if err := CheckAccessForPrivateNetwork("notarealuseraccount", "notarealnetwork", db); err == nil {
		t.Fatal("expected error")
	}
	// search for network user does not have access to
	if err := CheckAccessForPrivateNetwork("testuser", "thisnetworkdoesnotexist", db); err == nil {
		t.Fatal("expected error")
	}
	if err := CheckAccessForPrivateNetwork("testuser", "mynewnetworktotestwith", db); err != nil {
		t.Fatal(err)
	}
}

func Test_GetIPFSEndPoint(t *testing.T) {
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
	dev = !dev
	if net := api.GetIPFSEndpoint("network"); net == "" {
		t.Fatal("bad network url recovered")
	}
	dev = !dev
	if net := api.GetIPFSEndpoint("network"); net == "" {
		t.Fatal("bad network url recovered")
	}
}
