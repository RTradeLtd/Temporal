package queue

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	peer "gx/ipfs/QmcqU6QUDSXprb1518vYDGczrTJTyGwLG9eUa5iNX4xUtS/go-libp2p-peer"

	"github.com/RTradeLtd/Temporal/rtns"
	pb "github.com/RTradeLtd/grpc/krab"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/kaas"

	ci "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"

	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

type contextKey string

const (
	ipnsPublishTTL contextKey = "ipns-publish-ttl"
)

// ProcessIPNSEntryCreationRequests is used to process IPNS entry creation requests
func (qm *Manager) ProcessIPNSEntryCreationRequests(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	kb, err := kaas.NewClient(cfg.Endpoints)
	if err != nil {
		return err
	}
	// generate a temporary private key to reuse across our publisher
	pk, _, err := ci.GenerateKeyPair(ci.Ed25519, 256)
	if err != nil {
		return err
	}
	ipnsManager := models.NewIPNSManager(db)
	qm.LogInfo("processing ipns entry creation requests")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go func(d amqp.Delivery) {
				defer wg.Done()
				qm.LogInfo("neww message detected")
				ie := IPNSEntry{}
				err = json.Unmarshal(d.Body, &ie)
				if err != nil {
					qm.LogError(err, "failed to unmarshal message")
					d.Ack(false)
					return
				}
				// create our temporary publisher
				publisher, err := rtns.NewPublisher(&rtns.Opts{PK: pk}, "/ip4/0.0.0.0/tcp/3999")
				if err != nil {
					qm.refundCredits(ie.UserName, "ipns", ie.CreditCost, db)
					qm.LogError(err, "failed to build ipns publisher")
					d.Ack(false)
					return
				}
				if ie.NetworkName != "public" {
					qm.refundCredits(ie.UserName, "ipns", ie.CreditCost, db)
					qm.LogError(errors.New("private networks not supported"), "private networks not supported")
					d.Ack(false)
					return
				}
				qm.LogInfo("publishing ipns entry")
				// get the private key
				resp, err := kb.GetPrivateKey(context.Background(), &pb.KeyGet{Name: ie.Key})
				if err != nil {
					qm.refundCredits(ie.UserName, "ipns", ie.CreditCost, db)
					qm.LogError(err, "failed to retrieve private key")
					d.Ack(false)
					return
				}
				pk2, err := ci.UnmarshalPrivateKey(resp.PrivateKey)
				if err != nil {
					qm.refundCredits(ie.UserName, "ipns", ie.CreditCost, db)
					qm.LogError(err, "failed to unmarshal private key")
					d.Ack(false)
					return
				}
				eol := time.Now().Add(ie.LifeTime)
				// ignore the golint warning, as we must do this to pass in a ttl value
				// see https://discuss.ipfs.io/t/clarification-over-ttl-and-lifetime-for-ipns-records/4346 for more information
				ctx := context.WithValue(context.Background(), ipnsPublishTTL, ie.TTL)
				if err := publisher.PublishWithEOL(ctx, pk2, ie.CID, eol); err != nil {
					qm.refundCredits(ie.UserName, "ipns", ie.CreditCost, db)
					qm.LogError(err, "failed to publish ipns entry")
					d.Ack(false)
					return
				}
				id, err := peer.IDFromPrivateKey(pk2)
				if err != nil {
					// do not refund here since the record is published
					qm.LogError(err, "failed to unmarshal peer identity")
					d.Ack(false)
					return
				}
				if _, err := ipnsManager.FindByIPNSHash(id.Pretty()); err == nil {
					// if the previous equality check is true (err is nil) it means this entry already exists in the database
					if _, err = ipnsManager.UpdateIPNSEntry(
						id.Pretty(),
						ie.CID,
						ie.NetworkName,
						ie.UserName,
						ie.LifeTime,
						ie.TTL,
					); err != nil {
						qm.LogError(err, "failed to update ipns entry in database")
						d.Ack(false)
						return
					}
				} else {
					// record does not yet exist, so we must create a new one
					if _, err = ipnsManager.CreateEntry(
						id.Pretty(),
						ie.CID,
						ie.Key,
						ie.NetworkName,
						ie.UserName,
						ie.LifeTime,
						ie.TTL,
					); err != nil {
						qm.LogError(err, "failed to create ipns entry in database")
						d.Ack(false)
						return
					}
				}
				qm.LogInfo("successfully published and saved ipns entry")
				d.Ack(false)
				return // we must return here in order to trigger the wg.Done() defer
			}(d)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
			return nil
		}
	}
}
