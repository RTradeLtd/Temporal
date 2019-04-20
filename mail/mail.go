package mail

import (
	"errors"
	"sync"

	"github.com/RTradeLtd/config/v2"
	"github.com/RTradeLtd/database/v2/models"
	"github.com/RTradeLtd/gorm"
	"github.com/sendgrid/rest"
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// Mailer is a class that handles mail delivery
type Mailer interface {
	Send(email *mail.SGMailV3) (*rest.Response, error)
}

// Manager is our manager that handles email sending
type Manager struct {
	APIKey       string `json:"api_key"`
	EmailAddress string `json:"email_address"` // EmailAddress is the address from which messages will be sent from
	EmailName    string `json:"email_name"`    // EmailName is the name of the email address

	userManager *models.UserManager

	client Mailer
	cmux   sync.Mutex
}

// NewManager is used to create our mail manager, allowing us to send email
func NewManager(tCfg *config.TemporalConfig, db *gorm.DB) (*Manager, error) {
	var (
		apiKey       = tCfg.Sendgrid.APIKey
		emailAddress = tCfg.Sendgrid.EmailAddress
		emailName    = tCfg.Sendgrid.EmailName
		client       = sendgrid.NewSendClient(apiKey)
		um           = models.NewUserManager(db)
	)
	return &Manager{
		APIKey:       apiKey,
		EmailAddress: emailAddress,
		EmailName:    emailName,

		client:      client,
		userManager: um,
	}, nil
}

// BulkSend is used to handle sending a single email, to multiple recipients
func (mm *Manager) BulkSend(subject, content, contentType string, recipientNames, recipientEmails []string) error {
	if len(recipientNames) != len(recipientEmails) {
		return errors.New("recipientNames and recipientEmails must be fo equal length")
	}
	for k, v := range recipientEmails {
		if _, err := mm.SendEmail(subject, content, contentType, recipientNames[k], v); err != nil {
			return err
		}
	}
	return nil
}

// SendEmail is used to send an email to temporal users
func (mm *Manager) SendEmail(subject, content, contentType, recipientName, recipientEmail string) (int, error) {
	mm.cmux.Lock()
	if contentType == "" {
		contentType = "text/html"
	}

	var (
		from    = mail.NewEmail(mm.EmailName, mm.EmailAddress)
		to      = mail.NewEmail(recipientName, recipientEmail)
		message = mail.NewContent(contentType, content)
	)

	response, err := mm.client.Send(mail.NewV3MailInit(from, subject, to, message))
	mm.cmux.Unlock()
	if err != nil {
		return -1, err
	}
	return response.StatusCode, nil
}
