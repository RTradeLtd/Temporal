package queue

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/RTradeLtd/Temporal/mail"
	"github.com/RTradeLtd/gorm"
	"github.com/streadway/amqp"
)

// ProcessMailSends is a function used to process mail send queue messages
func (qm *Manager) ProcessMailSends(ctx context.Context, wg *sync.WaitGroup, db *gorm.DB, msgs <-chan amqp.Delivery) error {
	mm, err := mail.NewManager(qm.cfg, db)
	if err != nil {
		return err
	}
	qm.l.Info("processing email send requests")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processMailSend(d, wg, mm)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
			return nil
		}
	}
}

func (qm *Manager) processMailSend(d amqp.Delivery, wg *sync.WaitGroup, mm *mail.Manager) {
	defer wg.Done()
	qm.l.Info("new email send request detected")
	es := EmailSend{}
	if err := json.Unmarshal(d.Body, &es); err != nil {
		qm.l.Errorw(
			"failed to unmarshal message",
			"error", err.Error())
		d.Ack(false)
		return
	}
	for k, v := range es.Emails {
		_, err := mm.SendEmail(es.Subject, es.Content, es.ContentType, es.UserNames[k], v)
		if err != nil {
			qm.l.Errorw(
				"failed to send email",
				"error", err.Error(),
				"email", v,
				"user", es.UserNames[k])
		}
		qm.l.Infow(
			"email sent",
			"email", v,
			"user", es.UserNames[k])
	}
	d.Ack(false)
	return // we must return here in order to trigger the wg.Done() defer
}
