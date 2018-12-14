package queue

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	peer "gx/ipfs/QmY5Grm8pJdiSSVsYxx4uNRgweY72EmYwuSDbRnbFok3iY/go-libp2p-peer"

	"github.com/RTradeLtd/Temporal/rtns"
	pb "github.com/RTradeLtd/grpc/krab"

	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/kaas"

	ci "gx/ipfs/QmNiJiXwWE3kRhZrC5ej3kSjWHm337pYfhjLGSCDNKJP2s/go-libp2p-crypto"

	"github.com/streadway/amqp"
)

type contextKey string

const (
	ipnsPublishTTL contextKey = "ipns-publish-ttl"
)

// ProcessIPNSEntryCreationRequests is used to process IPNS entry creation requests
func (qm *Manager) ProcessIPNSEntryCreationRequests(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	kb, err := kaas.NewClient(qm.cfg.Endpoints)
	if err != nil {
		return err
	}
	// generate a temporary private key to reuse across our publisher
	pk, _, err := ci.GenerateKeyPair(ci.Ed25519, 256)
	if err != nil {
		return err
	}
	// user a long running publisher
	publisher, err := rtns.NewPublisher(&rtns.Opts{PK: pk}, "/ip4/0.0.0.0/tcp/3999")
	if err != nil {
		return err
	}
	ipnsManager := models.NewIPNSManager(qm.db)
	qm.LogInfo("processing ipns entry creation requests")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processIPNSEntryCreationRequest(d, wg, kb, publisher, ipnsManager)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
			return nil
		}
	}
}

func (qm *Manager) processIPNSEntryCreationRequest(d amqp.Delivery, wg *sync.WaitGroup, kb *kaas.Client, pub *rtns.Publisher, im *models.IpnsManager) {
	defer wg.Done()
	qm.LogInfo("neww message detected")
	ie := IPNSEntry{}
	if err := json.Unmarshal(d.Body, &ie); err != nil {
		qm.LogError(err, "failed to unmarshal message")
		d.Ack(false)
		return
	}
	// temporarily do not process ipns creation requests for non public networks
	if ie.NetworkName != "public" {
		qm.refundCredits(ie.UserName, "ipns", ie.CreditCost)
		qm.LogError(errors.New("private networks not supported"), "private networks not supported")
		d.Ack(false)
		return
	}
	qm.LogInfo("publishing ipns entry")
	// get the private key from krab to use with publishing
	resp, err := kb.GetPrivateKey(context.Background(), &pb.KeyGet{Name: ie.Key})
	if err != nil {
		qm.refundCredits(ie.UserName, "ipns", ie.CreditCost)
		qm.LogError(err, "failed to retrieve private key")
		d.Ack(false)
		return
	}
	// unmarshal the key that was returned by krab
	pk2, err := ci.UnmarshalPrivateKey(resp.PrivateKey)
	if err != nil {
		qm.refundCredits(ie.UserName, "ipns", ie.CreditCost)
		qm.LogError(err, "failed to unmarshal private key")
		d.Ack(false)
		return
	}
	// Note: context is used to pass in the experimental ttl value
	// see https://discuss.ipfs.io/t/clarification-over-ttl-and-lifetime-for-ipns-records/4346 for more information
	ctx := context.WithValue(context.Background(), ipnsPublishTTL, ie.TTL)
	eol := time.Now().Add(ie.LifeTime)
	if err := pub.PublishWithEOL(ctx, pk2, ie.CID, eol); err != nil {
		qm.refundCredits(ie.UserName, "ipns", ie.CreditCost)
		qm.LogError(err, "failed to publish ipns entry")
		d.Ack(false)
		return
	}
	// retrieve the peer id from the private key used to resolve the IPNS record
	id, err := peer.IDFromPrivateKey(pk2)
	if err != nil {
		// do not refund here since the record is published
		qm.LogError(err, "failed to unmarshal peer identity")
		d.Ack(false)
		return
	}
	// determine whether or not this ipns has been used, if so update record, otherwise create new one
	if _, err = im.FindByIPNSHash(id.Pretty()); err != nil {
		_, err = im.CreateEntry(id.Pretty(), ie.CID, ie.Key, ie.NetworkName, ie.UserName, ie.LifeTime, ie.TTL)
	} else {
		_, err = im.UpdateIPNSEntry(id.Pretty(), ie.CID, ie.NetworkName, ie.UserName, ie.LifeTime, ie.TTL)
	}
	if err != nil {
		qm.LogError(err, "failed to save ipns entry in database, but the record was published")
	} else {
		qm.LogInfo("successfully published and saved ipns entry to database")
	}
	d.Ack(false)
	return // we must return here in order to trigger the wg.Done() defer

}
