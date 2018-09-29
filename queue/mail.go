package queue

import (
	"encoding/json"

	"github.com/RTradeLtd/Temporal/mail"
	"github.com/RTradeLtd/config"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var (
	// IpfsPinFailedContent is a to-be formatted message sent on IPFS pin failures
	IpfsPinFailedContent = "Pin failed for content hash %s on IPFS network %s, for reason %s"
	// IpfsPinFailedSubject is a subject for IPFS pin failed messages
	IpfsPinFailedSubject = "IPFS Pin Failed"
	// IpfsFileFailedContent is a to be formatted message sent on ipfs add failures
	IpfsFileFailedContent = "IPFS File Add Failed for object name %s on IPFS network %s"
	// IpfsFileFailedSubject is a subject for ipfs file add fails
	IpfsFileFailedSubject = "IPFS File Add Failed"
	// IpfsPrivateNetworkUnauthorizedSubject is a subject whenever someone tries to access a bad private network
	IpfsPrivateNetworkUnauthorizedSubject = "Unauthorized access to IPFS private network"
	// IpfsInitializationFailedSubject is a subject used when connecting to ipfs fails
	IpfsInitializationFailedSubject = "Connection to IPFS failed"
	// IpnsEntryFailedSubject is a subject sent upon IPNS failures
	IpnsEntryFailedSubject = "IPNS Entry Creation Failed"
	// IpnsEntryFailedContent is the content used when sending an email for IPNS entry creation failures
	IpnsEntryFailedContent = "IPNS Entry creation failed for content hash %s using key %s for reason %s"
	// PaymentConfirmationFailedSubject is a subject used when payment confirmations fail
	PaymentConfirmationFailedSubject = "Payment Confirmation Failed"
	// PaymentConfirmationFailedContent is a content used when a payment confirmation failure occurs
	PaymentConfirmationFailedContent = "Payment failed for content hash %s with error %s"
)

// EmailSend is a helper struct used to contained formatted content ot send as an email
type EmailSend struct {
	Subject     string   `json:"subject"`
	Content     string   `json:"content"`
	ContentType string   `json:"content_type"`
	UserNames   []string `json:"user_names"`
}

// ProcessMailSends is a function used to process mail send queue messages
func (qm *QueueManager) ProcessMailSends(msgs <-chan amqp.Delivery, tCfg *config.TemporalConfig) error {
	mm, err := mail.GenerateMailManager(tCfg)
	if err != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"error":   err.Error(),
		}).Error("failed to generate mail manager")
		return err
	}
	qm.Logger.WithFields(log.Fields{
		"service": qm.QueueName,
	}).Info("process email sends")
	for d := range msgs {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("detected new message")
		es := EmailSend{}
		err = json.Unmarshal(d.Body, &es)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   err.Error(),
			}).Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}
		emails := make(map[string]string)
		for _, v := range es.UserNames {
			resp, err := mm.UserManager.FindEmailByUserName(v)
			if err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"user":    v,
					"error":   err.Error(),
				}).Error("failed to find email by user name")
				d.Ack(false)
				continue
			}
			emails[v] = resp[v]
		}
		for k, v := range emails {
			_, err = mm.SendEmail(es.Subject, es.Content, es.ContentType, k, v)
			if err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"user":    k,
					"error":   err.Error(),
				}).Error("failed to send email")
			}
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"users":   es.UserNames,
		}).Info("successfully sent emails")
		d.Ack(false)
	}
	return nil
}
