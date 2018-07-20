package mail

import (
	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/sendgrid/sendgrid-go"
)

/*
This package is used to handle sending email messages to
TEMPORAL users
*/

type MailManager struct {
	APIKey       string              `json:"api_key"`
	EmailAddress string              `json:"email_address"` // EmailAddress is the address from which messages will be sent from
	EmailName    string              `json:"email_name"`    // EmailName is the name of the email address
	Client       *sendgrid.Client    `json:"client"`
	UserManager  *models.UserManager `json:"user_manager"`
}

func GenerateMailManager(tCfg *config.TemporalConfig) (*MailManager, error) {
	apiKey := tCfg.Sendgrid.APIKey
	emailAddress := tCfg.Sendgrid.EmailAddress
	emailName := tCfg.Sendgrid.EmailName
	client := sendgrid.NewSendClient(apiKey)

	dbPass := tCfg.Database.Password
	dbURL := tCfg.Database.URL
	dbUser := tCfg.Database.Username
	db, err := database.OpenDBConnection(dbPass, dbURL, dbUser)
	if err != nil {
		return nil, err
	}
	um := models.NewUserManager(db)
	mm := MailManager{
		APIKey:       apiKey,
		EmailAddress: emailAddress,
		EmailName:    emailName,
		Client:       client,
		UserManager:  um,
	}
	return &mm, nil
}

/*
// Send24HourEmail is a function used to send report information for the last 24 hour period
func (m *Manager) Send24HourEmail(date, ethMined, usdValue, cadValue string) (int, error) {
	content := fmt.Sprintf("<br>Eth Mined: %v<br>USD Value: %v<br>CAD Value: %v", ethMined, usdValue, cadValue)
	from := mail.NewEmail("stake-sendgrid-api", "sgapi@rtradetechnologies.com")
	subject := fmt.Sprintf("Ethereum Mining Report - %s", date)
	to := mail.NewEmail("Mining Reports", "postables@rtradetechnologies.com")

	mContent := mail.NewContent("text/html", content)
	mail := mail.NewV3MailInit(from, subject, to, mContent)

	response, err := m.SendgridClient.Send(mail)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
*/
