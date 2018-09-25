package queue

import (
	"encoding/json"

	"github.com/RTradeLtd/Temporal/mail"
	"github.com/RTradeLtd/config"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

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
