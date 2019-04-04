package v3

import "github.com/RTradeLtd/database/models"

type userManager interface {
	NewUserAccount(username, password, email string) (*models.User, error)
	SignIn(username, password string) (bool, error)
	FindByUserName(username string) (*models.User, error)

	GenerateEmailVerificationToken(username string) (*models.User, error)
	ValidateEmailVerificationToken(username, challenge string) (*models.User, error)
}

type usageManager interface {
	NewUsageEntry(username string, tier models.DataUsageTier) (*models.Usage, error)
	FindByUserName(username string) (*models.Usage, error)
}

type publisher interface {
	PublishMessage(body interface{}) error
}
