package queue

import (
	"encoding/json"
	"time"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/rtfs"
	"github.com/RTradeLtd/rtfs/krab"

	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessIPNSEntryCreationRequests is used to process IPNS entry creation requests
func (qm *Manager) ProcessIPNSEntryCreationRequests(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	kb, err := krab.NewKrab(krab.Opts{Passphrase: cfg.IPFS.KrabPassword, DSPath: cfg.IPFS.KeystorePath, ReadOnly: true})
	if err != nil {
		return err
	}
	keystore, err := rtfs.NewKeystoreManager(kb)
	if err != nil {
		return err
	}
	ipfsManager, err := rtfs.NewManager(
		cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port,
		keystore,
		time.Minute*10,
	)
	if err != nil {
		return err
	}
	ipnsManager := models.NewIPNSManager(db)
	userManager := models.NewUserManager(db)
	networkManager := models.NewHostedIPFSNetworkManager(db)
	qm.LogInfo("processing ipns entry creation requests")
	for d := range msgs {
		qm.LogInfo("neww message detected")
		ie := IPNSEntry{}
		err = json.Unmarshal(d.Body, &ie)
		if err != nil {
			qm.LogError(err, "failed to unmarshal message")
			d.Ack(false)
			continue
		}
		apiURL := ""
		if ie.NetworkName != "public" {
			canAccess, err := userManager.CheckIfUserHasAccessToNetwork(ie.UserName, ie.NetworkName)
			if err != nil {
				qm.refundCredits(ie.UserName, "ipns", ie.CreditCost, db)
				qm.LogError(err, "failed to check for private network access")
				d.Ack(false)
				continue
			}
			if !canAccess {
				qm.refundCredits(ie.UserName, "ipns", ie.CreditCost, db)
				qm.LogError(err, "invalid private network access")
				d.Ack(false)
				continue
			}
			apiURLName, err := networkManager.GetAPIURLByName(ie.NetworkName)
			if err != nil {
				qm.refundCredits(ie.UserName, "ipns", ie.CreditCost, db)
				qm.LogError(err, "failed to find ipfs api url")
				d.Ack(false)
				continue
			}
			apiURL = apiURLName
			qm.LogInfo("initializing connection to private ipfs network")
			ipfsManager, err = rtfs.NewManager(apiURL, keystore, time.Minute*10)
			if err != nil {
				qm.refundCredits(ie.UserName, "ipns", ie.CreditCost, db)
				qm.LogError(err, "failed to initialized connection to private ifps network")
				d.Ack(false)
				continue
			}
		}
		qm.LogInfo("publishing ipns entry")
		response, err := ipfsManager.Publish(ie.CID, ie.Key, ie.LifeTime, ie.TTL, ie.Resolve)
		if err != nil {
			qm.refundCredits(ie.UserName, "ipns", ie.CreditCost, db)
			qm.LogError(err, "failed to publish ipns entry")
			d.Ack(false)
			continue
		}
		if _, err := ipnsManager.FindByIPNSHash(response.Name); err == nil {
			// if the previous equality check is true (err is nil) it means this entry already exists in the database
			if _, err = ipnsManager.UpdateIPNSEntry(
				response.Name,
				ie.CID,
				ie.NetworkName,
				ie.UserName,
				ie.LifeTime,
				ie.TTL,
			); err != nil {
				qm.LogError(err, "failed to update ipns entry in database")
				d.Ack(false)
				continue
			}
		} else {
			// record does not yet exist, so we must create a new one
			if _, err = ipnsManager.CreateEntry(
				response.Name,
				ie.CID,
				ie.Key,
				ie.NetworkName,
				ie.UserName,
				ie.LifeTime,
				ie.TTL,
			); err != nil {
				qm.LogError(err, "failed to create ipns entry in database")
				d.Ack(false)
				continue
			}
		}
		qm.LogInfo("successfully published and saved ipns entry")
		d.Ack(false)
	}
	return nil
}
