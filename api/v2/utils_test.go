package v2

import (
	"testing"

	"github.com/RTradeLtd/Temporal/mocks"
	"github.com/RTradeLtd/Temporal/utils"
	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2/models"
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
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	fakeWalletService := &mocks.FakeWalletServiceClient{}

	api, err := setupAPI(t, fakeLens, fakeOrch, fakeSigner, fakeWalletService, cfg, db)
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
	fakeLens := &mocks.FakeLensV2Client{}
	fakeOrch := &mocks.FakeServiceClient{}
	fakeSigner := &mocks.FakeSignerClient{}
	fakeWalletService := &mocks.FakeWalletServiceClient{}

	api, err := setupAPI(t, fakeLens, fakeOrch, fakeSigner, fakeWalletService, cfg, db)
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

func Test_Ensure_Two_Year_Max(t *testing.T) {
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
	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	um := models.NewUploadManager(db)
	upload, err := um.NewUpload(
		randString,
		"file",
		models.UploadOptions{
			Username:         "testuser",
			NetworkName:      "public",
			HoldTimeInMonths: 1,
			Encrypted:        false,
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	// TODO(bonedaddy): properly test
	type args struct {
		holdTimeInMonths int64
		upload           *models.Upload
		isFree           bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"12-Months-paid", args{12, upload, false}, false},
		{"22-Months-paid", args{22, upload, false}, false},
		{"25-Months-paid", args{25, upload, false}, true},
		{"12-Months-free", args{12, upload, true}, false},
		{"22-Months-free", args{22, upload, true}, false},
		{"25-Months-free", args{25, upload, true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := api.ensureLEMaxPinTime(
				tt.args.upload,
				tt.args.holdTimeInMonths,
				tt.args.isFree,
			); (err != nil) != tt.wantErr {
				t.Fatalf("ensureTwoYearMax err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
