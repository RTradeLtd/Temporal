package mail

import (
	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/database"
	"github.com/RTradeLtd/Temporal/models"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
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

// SendEmail is used to send an email to temporal users
func (mm *MailManager) SendEmail(subject, content, contentType, recipientName, recipientEmail string) (int, error) {
	if contentType == "" {
		contentType = "text/html"
	}
	// 	content := fmt.Sprintf("<br>Eth Mined: %v<br>USD Value: %v<br>CAD Value: %v", ethMined, usdValue, cadValue)
	from := mail.NewEmail(mm.EmailName, mm.EmailAddress)
	to := mail.NewEmail(recipientName, recipientEmail)

	mContent := mail.NewContent(contentType, content)
	mail := mail.NewV3MailInit(from, subject, to, mContent)

	response, err := mm.Client.Send(mail)
	if err != nil {
		return 0, err
	}
	return response.StatusCode, nil
}
