package queue

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	ci "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
	"github.com/streadway/amqp"

	"github.com/RTradeLtd/database/v2/models"
	pb "github.com/RTradeLtd/grpc/krab"
	kaas "github.com/RTradeLtd/kaas/v2"
	"github.com/RTradeLtd/rtns"
	"github.com/multiformats/go-multiaddr"
)

type contextKey string

const (
	ipnsPublishTTL contextKey = "ipns-publish-ttl"
)

func (qm *Manager) getPublisherKey(ctx context.Context, kb *kaas.Client) (ci.PrivKey, error) {
	// generate a temporary key if we have not specified a key name
	if qm.cfg.Services.RTNS.KeyName == "" {
		pk, _, err := ci.GenerateKeyPair(ci.Ed25519, 256)
		return pk, err
	}
	resp, err := kb.GetPrivateKey(ctx, &pb.KeyGet{Name: qm.cfg.Services.RTNS.KeyName})
	if err != nil {
		return nil, err
	}
	return ci.UnmarshalPrivateKey(resp.GetPrivateKey())
}

// ProcessIPNSEntryCreationRequests is used to process IPNS entry creation requests
func (qm *Manager) ProcessIPNSEntryCreationRequests(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	kbPrimary, err := kaas.NewClient(qm.cfg.Services, false)
	if err != nil {
		return err
	}
	kbBackup, err := kaas.NewClient(qm.cfg.Services, true)
	if err != nil {
		return err
	}
	pubPK, err := qm.getPublisherKey(ctx, kbPrimary)
	addr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/3999")
	if err != nil {
		return err
	}
	rConfig := rtns.Config{
		PK:          pubPK,
		ListenAddrs: []multiaddr.Multiaddr{addr},
		DSPath:      qm.cfg.Services.RTNS.DatastorePath,
	}
	publisher, err := rtns.NewRTNS(ctx, kbPrimary, rConfig)
	if err != nil {
		return err
	}
	publisher.DefaultBootstrap()
	ipnsManager := models.NewIPNSManager(qm.db)
	qm.l.Info("processing ipns entry creation requests")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processIPNSEntryCreationRequest(d, wg, kbBackup, publisher, ipnsManager)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
			return nil
		case msg := <-qm.ErrCh:
			qm.Close()
			wg.Done()
			qm.l.Errorw(
				"a protocol connection error stopping rabbitmq was received",
				"error", msg.Error())
			return errors.New(ErrReconnect)
		}
	}
}

func (qm *Manager) processIPNSEntryCreationRequest(d amqp.Delivery, wg *sync.WaitGroup, kbBackup *kaas.Client, pub *rtns.RTNS, im *models.IpnsManager) {
	defer wg.Done()
	qm.l.Info("new ipns entry creation detected")
	ie := IPNSEntry{}
	if err := json.Unmarshal(d.Body, &ie); err != nil {
		qm.l.Errorw(
			"failed to unmarshal message",
			"error", err.Error())
		d.Ack(false)
		return
	}
	// temporarily do not process ipns creation requests for non public networks
	if ie.NetworkName != "public" {
		qm.l.Errorw(
			"private networks not supported for ipns",
			"user", ie.UserName)
		d.Ack(false)
		return
	}
	qm.l.Infow(
		"publishing ipns entry",
		"user", ie.UserName,
		"key", ie.Key,
		"cid", ie.CID)
	var cache bool
	// first attempt to retrieve private key
	// from the primary krab keystore which
	// is embedded into the RTNS publisher
	pk, err := pub.Keys.Get(ie.Key)
	if err != nil {
		// do not cache entries if the key is not available in the primary keystore
		cache = false
		qm.l.Warnw(
			"failed to retrieve private key from priamry krab, attempting backup",
			"error", err.Error(),
			"user", ie.UserName,
			"key", ie.Key,
			"cid", ie.CID,
		)
		if !qm.dev {
			var errCheck error
			resp, errCheck := kbBackup.GetPrivateKey(context.Background(), &pb.KeyGet{Name: ie.Key})
			if errCheck != nil {
				qm.refundCredits(ie.UserName, "ipns", ie.CreditCost)
				qm.l.Errorw(
					"failed to retrieve private key from backup krab",
					"error", err.Error(),
					"user", ie.UserName,
					"key", ie.Key,
					"cid", ie.CID)
				d.Ack(false)
				return
			}
			pk, err = ci.UnmarshalPrivateKey(resp.GetPrivateKey())
			if err != nil {
				qm.refundCredits(ie.UserName, "ipns", ie.CreditCost)
				qm.l.Errorw(
					"failed to unmarshal private key",
					"error", err.Error(),
					"user", ie.UserName,
					"key", ie.Key,
					"cid", ie.CID)
				d.Ack(false)
				return
			}
		} else {
			qm.l.Errorw(
				"primary krab key retrieval failure, with dev mode disabled, aborting",
				"user", ie.UserName,
				"key", ie.Key,
				"cid", ie.CID,
			)
			d.Ack(false)
			return
		}
	} else {
		// only cache entries whose key is available via the primary keystore
		cache = true
	}
	// Note: context is used to pass in the experimental ttl value
	// see https://discuss.ipfs.io/t/clarification-over-ttl-and-lifetime-for-ipns-records/4346 for more information
	ctx := context.WithValue(context.Background(), ipnsPublishTTL, ie.TTL)
	eol := time.Now().Add(ie.LifeTime)
	if err := pub.PublishWithEOL(ctx, pk, eol, cache, ie.Key, ie.CID); err != nil {
		qm.refundCredits(ie.UserName, "ipns", ie.CreditCost)
		qm.l.Errorw(
			"failed to publish ipns entry",
			"error", err.Error(),
			"user", ie.UserName,
			"key", ie.Key,
			"cid", ie.CID)
		d.Ack(false)
		return
	}
	// retrieve the peer id from the private key used to resolve the IPNS record
	id, err := peer.IDFromPrivateKey(pk)
	if err != nil {
		// do not refund here since the record is published
		qm.l.Errorw(
			"failed to unmarshal peer identity private key",
			"error", err.Error(),
			"user", ie.UserName,
			"key", ie.Key,
			"cid", ie.CID)
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
		qm.l.Errorw(
			"failed to update ipns entry in database",
			"error", err.Error(),
			"user", ie.UserName,
			"key", ie.Key,
			"cid", ie.CID)
	} else {
		qm.l.Infow(
			"successfully processed ipns entry creation request",
			"user", ie.UserName,
			"key", ie.Key,
			"cid", ie.CID)
	}
	d.Ack(false)
	return // we must return here in order to trigger the wg.Done() defer

}
