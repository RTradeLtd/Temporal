package queue

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database/models"
	log "github.com/sirupsen/logrus"

	"github.com/RTradeLtd/Temporal/rtfs"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// IPNSEntry is used to hold relevant information needed to process IPNS entry creation requests
type IPNSEntry struct {
	CID         string        `json:"cid"`
	LifeTime    time.Duration `json:"life_time"`
	TTL         time.Duration `json:"ttl"`
	Resolve     bool          `json:"resolve"`
	Key         string        `json:"key"`
	UserName    string        `json:"user_name"`
	NetworkName string        `json:"network_name"`
}

// ProcessIPNSEntryCreationRequests is used to process IPNS entry creation requests
func (qm *QueueManager) ProcessIPNSEntryCreationRequests(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	ipfsManager, err := rtfs.Initialize("", "")
	if err != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"error":   err.Error(),
		}).Error("failed to initialize connection to ipfs")
		return err
	}
	err = ipfsManager.CreateKeystoreManager()
	if err != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"error":   err.Error(),
		}).Error("failed to create keystore manager")
		return err
	}
	ipnsManager := models.NewIPNSManager(db)
	userManager := models.NewUserManager(db)
	networkManager := models.NewHostedIPFSNetworkManager(db)
	qmEmail, err := Initialize(IpnsEntryQueue, cfg.RabbitMQ.URL, true, false)
	if err != nil {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"error":   err.Error(),
		}).Error("failed to initialize connection to email send queue")
		return err
	}
	qm.Logger.WithFields(log.Fields{
		"service": qm.QueueName,
	}).Info("Processing ipns entry requests")
	for d := range msgs {
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
		}).Info("new message detected")
		ie := IPNSEntry{}
		err = json.Unmarshal(d.Body, &ie)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"error":   err.Error(),
			}).Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}
		apiURL := ""
		if ie.NetworkName != "public" {
			canAccess, err := userManager.CheckIfUserHasAccessToNetwork(ie.UserName, ie.NetworkName)
			if err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"user":    ie.UserName,
					"network": ie.NetworkName,
					"error":   err.Error(),
				}).Error("error checking for private network access")
				d.Ack(false)
				continue
			}
			if !canAccess {
				addresses := []string{}
				addresses = append(addresses, ie.UserName)
				es := EmailSend{
					Subject:     IpfsPrivateNetworkUnauthorizedSubject,
					Content:     fmt.Sprintf("Unauthorized access to IPFS private network %s", ie.NetworkName),
					ContentType: "",
					UserNames:   addresses,
				}
				err = qmEmail.PublishMessage(es)
				if err != nil {
					qm.Logger.WithFields(log.Fields{
						"service": qm.QueueName,
						"error":   err.Error(),
					}).Error("failed to publish message to email send queue")
				}
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"user":    ie.UserName,
					"network": ie.NetworkName,
				}).Error("unauthorized access to private network")
				d.Ack(false)
				continue
			}
			apiURLName, err := networkManager.GetAPIURLByName(ie.NetworkName)
			if err != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"user":    ie.UserName,
					"network": ie.NetworkName,
					"error":   err.Error(),
				}).Error("failed to get ipfs api url by name")
				d.Ack(false)
				continue
			}
			apiURL = apiURLName
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"user":    ie.UserName,
				"network": ie.NetworkName,
			}).Info("initializing connection to private ipfs network")
			ipfsManager, err = rtfs.Initialize("", apiURL)
			if err != nil {
				addresses := []string{}
				addresses = append(addresses, ie.UserName)
				es := EmailSend{
					Subject:     IpfsInitializationFailedSubject,
					Content:     fmt.Sprintf("Connection to IPFS failed due to the following error %s", err),
					ContentType: "",
					UserNames:   addresses,
				}
				errOne := qmEmail.PublishMessage(es)
				if errOne != nil {
					qm.Logger.WithFields(log.Fields{
						"service": qm.QueueName,
						"error":   errOne.Error(),
					}).Error("failed to publish message to email send queue")
				}
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"user":    ie.UserName,
					"network": ie.NetworkName,
					"error":   err.Error(),
				}).Error("failed to initialize conenction to private ipfs network")
				d.Ack(false)
				continue
			}
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"user":    ie.UserName,
			"network": ie.NetworkName,
		}).Info("publishing ipns entry")
		response, err := ipfsManager.PublishToIPNSDetails(ie.CID, ie.Key, ie.LifeTime, ie.TTL, ie.Resolve)
		if err != nil {
			fmt.Println("error publishing response")
			formattedContent := fmt.Sprintf(IpnsEntryFailedContent, ie.CID, ie.Key, err)
			addresses := []string{}
			addresses = append(addresses, ie.UserName)
			es := EmailSend{
				Subject:     IpnsEntryFailedSubject,
				Content:     formattedContent,
				ContentType: "",
				UserNames:   addresses,
			}
			errOne := qmEmail.PublishMessage(es)
			if errOne != nil {
				qm.Logger.WithFields(log.Fields{
					"service": qm.QueueName,
					"error":   errOne.Error(),
				}).Error("failed to publish message to email send queue")
			}
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"user":    ie.UserName,
				"network": ie.NetworkName,
				"error":   err.Error(),
			}).Error("failed to publish entry to ipns")
			d.Ack(false)
			continue
		}
		_, err = ipnsManager.UpdateIPNSEntry(response.Name, ie.CID, ie.Key, ie.NetworkName, ie.LifeTime, ie.TTL)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.QueueName,
				"user":    ie.UserName,
				"network": ie.NetworkName,
				"error":   err.Error(),
			}).Error("failed to update IPNS entry in database")
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.QueueName,
			"user":    ie.UserName,
			"network": ie.NetworkName,
		}).Info("successfully published entry to ipns")
		d.Ack(false)
	}
	return nil
}
