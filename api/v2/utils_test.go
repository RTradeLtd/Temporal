package v2

import (
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/RTradeLtd/config"
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
	valid, err := api.verifyEmailJWTToken(tkn, randUser)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Fatal("failed to verify the jwt")
	}
}
