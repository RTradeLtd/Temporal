package queue

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/RTradeLtd/Temporal/mail"
	"github.com/streadway/amqp"
)

// ProcessMailSends is a function used to process mail send queue messages
func (qm *Manager) ProcessMailSends(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	mm, err := mail.NewManager(qm.cfg)
	if err != nil {
		qm.LogError(err, "failed to generate mail manager")
		return err
	}
	qm.LogInfo("processing email sends")
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
	qm.LogInfo("detected new message")
	es := EmailSend{}
	if err := json.Unmarshal(d.Body, &es); err != nil {
		qm.LogError(err, "failed to unmarshal message")
		d.Ack(false)
		return
	}
	var emailSent bool
	for k, v := range es.Emails {
		_, err := mm.SendEmail(es.Subject, es.Content, es.ContentType, es.UserNames[k], v)
		if err != nil {
			qm.LogError(err, "failed to send email")
			d.Ack(false)
			emailSent = false
			continue
		}
		emailSent = true
	}
	if !emailSent {
		return
	}
	qm.LogInfo("successfully sent email(s)")
	d.Ack(false)
	return // we must return here in order to trigger the wg.Done() defer
}
