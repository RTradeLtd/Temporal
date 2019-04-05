package v3

import "github.com/RTradeLtd/database/models"

type userManager interface {
	FindByUserName(username string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)

	NewUserAccount(username, password, email string) (*models.User, error)
	SignIn(username, password string) (bool, error)

	ResetPassword(username string) (string, error)
	ChangePassword(username string, curr string, new string) (bool, error)
	UpdateCustomerObjectHash(username, newHash string) error

	GenerateEmailVerificationToken(username string) (*models.User, error)
	ValidateEmailVerificationToken(username, challenge string) (*models.User, error)
}

type creditsManager interface {
	AddCredits(username string, credits float64) (*models.User, error)
	RemoveCredits(username string, credits float64) (*models.User, error)
}

type usageManager interface {
	NewUsageEntry(username string, tier models.DataUsageTier) (*models.Usage, error)
	FindByUserName(username string) (*models.Usage, error)
	UpdateTier(username string, tier models.DataUsageTier) error
}

type publisher interface {
	PublishMessage(body interface{}) error
}
