package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/RTradeLtd/Temporal/log"
	"github.com/RTradeLtd/kaas"
	"github.com/RTradeLtd/rtfs"

	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/gorm"
	"github.com/streadway/amqp"

	pb "github.com/RTradeLtd/grpc/krab"
	ci "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

// ProcessIPFSKeyCreation is used to create IPFS keys
func (qm *Manager) ProcessIPFSKeyCreation(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	kb, err := kaas.NewClient(qm.cfg.Endpoints)
	if err != nil {
		return err
	}
	userManager := models.NewUserManager(qm.db)
	qm.l.Info("processing ipfs key creation requests")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processIPFSKeyCreation(d, wg, kb, userManager)
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

// ProccessIPFSPins is used to process IPFS pin requests
func (qm *Manager) ProccessIPFSPins(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	userManager := models.NewUserManager(qm.db)
	networkManager := models.NewHostedIPFSNetworkManager(qm.db)
	uploadManager := models.NewUploadManager(qm.db)
	logger, err := log.NewLogger(qm.cfg.LogDir+"cluster_publisher.log", false)
	if err != nil {
		return err
	}
	// initialize a connection to the cluster pin queue so we can trigger pinning of this content to our cluster
	qmCluster, err := New(IpfsClusterPinQueue, qm.cfg.RabbitMQ.URL, true, logger)
	if err != nil {
		qm.l.Errorw("failed to intialize cluster pin queue connection", "error", err.Error())
		return err
	}
	ipfsManager, err := rtfs.NewManager(qm.cfg.IPFS.APIConnection.Host+":"+qm.cfg.IPFS.APIConnection.Port, "", time.Minute*60, false)
	if err != nil {
		qm.l.Errorw("failed to initialize connection to ipfs", "error", err.Error())
		return err
	}
	qm.l.Info("processing ipfs pins")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processIPFSPin(d, wg, userManager, networkManager, uploadManager, qmCluster, ipfsManager)
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

func (qm *Manager) processIPFSPin(d amqp.Delivery, wg *sync.WaitGroup, usrm *models.UserManager, nm *models.IPFSNetworkManager, upldm *models.UploadManager, qmCluster *Manager, ipfsManager *rtfs.IpfsManager) {
	defer wg.Done()
	qm.l.Info("new pin request detected")
	pin := &IPFSPin{}
	if err := json.Unmarshal(d.Body, pin); err != nil {
		qm.l.Errorw("failed to unmarshal message", "error", err.Error())
		d.Ack(false)
		return
	}
	// setup the default api connection
	apiURL := qm.cfg.IPFS.APIConnection.Host + ":" + qm.cfg.IPFS.APIConnection.Port
	// check whether or not this pin is for a private network
	// if it is, verify whether the user has acess to the network, and retrieve the api url
	if pin.NetworkName != "public" {
		canAccess, err := usrm.CheckIfUserHasAccessToNetwork(pin.UserName, pin.NetworkName)
		if err != nil {
			qm.l.Errorw("failed to lookup private network in database", "error", err.Error())
			d.Ack(false)
			return
		}
		if !canAccess {
			qm.l.Errorw(
				"unauthorized private network access",
				"error", errors.New("user does not have access to private network").Error(),
				"user", pin.UserName)
			d.Ack(false)
			return
		}
		apiURL = fmt.Sprintf("%s/network/%s/api", qm.cfg.Nexus.Host+":"+qm.cfg.Nexus.Delegator.Port, pin.NetworkName)
		// connect to ipfs
		ipfsManager, err = rtfs.NewManager(apiURL, pin.JWT, time.Minute*10, true)
		if err != nil {
			qm.l.Infow(
				"failed to initialize connection to ipfs",
				"error", err.Error(),
				"user", pin.UserName,
				"network", pin.NetworkName)
			d.Ack(false)
			return
		}
	}
	qm.l.Infow(
		"initializing connection to ipfs",
		"user", pin.UserName)
	qm.l.Infow(
		"pinning hash to ipfs",
		"cid", pin.CID,
		"user", pin.UserName,
		"network", pin.NetworkName)
	// pin the content
	if err := ipfsManager.Pin(pin.CID); err != nil {
		if pin.NetworkName == "public" {
			qm.refundCredits(pin.UserName, "pin", pin.CreditCost)
		}
		models.NewUsageManager(qm.db).ReduceDataUsage(pin.UserName, uint64(pin.Size))
		qm.l.Errorw(
			"failed to pin hash to ipfs",
			"error", err.Error(),
			"user", pin.UserName,
			"network", pin.NetworkName)
		d.Ack(false)
		return
	}
	// cluster support for private networks isn't available yet
	// as such, skip additional processing for cluster pins
	qm.l.Infof(
		"successfully process pin request",
		"user", pin.UserName,
		"network", pin.NetworkName)
	upload, err := upldm.FindUploadByHashAndNetwork(pin.CID, pin.NetworkName)
	if err != nil && err != gorm.ErrRecordNotFound {
		qm.l.Errorw(
			"fail to check database for upload",
			"error", err.Error(),
			"user", pin.UserName)
		d.Ack(false)
		return
	}
	// check whether or not we have seen this content hash before to determine how database needs to be updated
	if upload == nil {
		_, err = upldm.NewUpload(pin.CID, "pin", models.UploadOptions{
			NetworkName:      pin.NetworkName,
			Username:         pin.UserName,
			HoldTimeInMonths: pin.HoldTimeInMonths})
	} else {
		// the record already exists so we will update
		_, err = upldm.UpdateUpload(pin.HoldTimeInMonths, pin.UserName, pin.CID, pin.NetworkName)
	}
	// validate whether or not the database was updated properly
	if err != nil {
		qm.l.Errorw(
			"failed to update database",
			"error", err.Error(),
			"user", pin.UserName)
	}
	d.Ack(false)
	return // we must return here in order to trigger the wg.Done() defer
}

func (qm *Manager) processIPFSKeyCreation(d amqp.Delivery, wg *sync.WaitGroup, kb *kaas.Client, um *models.UserManager) {
	defer wg.Done()
	qm.l.Info("new key creation request detected")
	key := IPFSKeyCreation{}
	if err := json.Unmarshal(d.Body, &key); err != nil {
		qm.l.Errorw(
			"failed to unmarshal message",
			"error", err.Error())
		d.Ack(false)
		return
	}
	// to prevent key name collision, we need to ensure that the keyname was prefixed with their username and a hyphen
	// whenever a user creates a key, the API call will prepend their username and a hyphen before sending the message for processing
	// this check ensures that the key was properly prefixed
	if strings.Split(key.Name, "-")[0] != key.UserName {
		qm.l.Errorf("invalid key name %s, must be prefixed with: %s-", key.Name, key.UserName)
		d.Ack(false)
		return
	}
	var (
		keyTypeInt int
		bitsInt    int
	)
	// validate the key parameters used for creation
	switch key.Type {
	case "rsa":
		keyTypeInt = ci.RSA
		// ensure the provided key size is within a valid range, otherwise default to 2048
		if key.Size > 4096 || key.Size < 2048 {
			bitsInt = 2048
		} else {
			bitsInt = key.Size
		}
	case "ed25519":
		keyTypeInt = ci.Ed25519
		// ed25519 keys use 256 bits, so regardless of what the user provides for bit size, hard set 256
		bitsInt = 256
	default:
		qm.l.Errorw(
			"invalid key type for creation request",
			"error", fmt.Errorf("key must be ed25519 or rsa, not %s", key.Type),
			"user", key.UserName,
			"key_name", key.Name)
		d.Ack(false)
		return
	}
	// generate the appropriate keypair
	pk, _, err := ci.GenerateKeyPair(keyTypeInt, bitsInt)
	if err != nil {
		qm.l.Errorw(
			"failed to create key",
			"error", err.Error(),
			"user", key.UserName,
			"key_name", key.Name)
		d.Ack(false)
		return
	}
	// retrieve a human friendly format, also verifying the key is a valid ipfs key
	id, err := peer.IDFromPrivateKey(pk)
	if err != nil {
		qm.l.Errorw(
			"failed to get peer id from private key",
			"error", err.Error(),
			"user", key.UserName,
			"Key_name", key.Name)
		d.Ack(false)
		return
	}
	// convert the key to bytes to send to krab for processing
	pkBytes, err := pk.Bytes()
	if err != nil {
		qm.l.Errorw(
			"failed to marshal key to bytes",
			"error", err.Error(),
			"user", key.UserName,
			"key_name", key.Name)
		d.Ack(false)
		return
	}

	// store the key in krab
	if _, err := kb.PutPrivateKey(context.Background(), &pb.KeyPut{Name: key.Name, PrivateKey: pkBytes}); err != nil {
		qm.l.Errorw(
			"failed to store key in krab",
			"error", err.Error(),
			"user", key.UserName,
			"key_name", key.Name)
		d.Ack(false)
		return
	}
	// doesn't need a refund, key was generated and stored in our keystore, but information not saved to db
	if err := um.AddIPFSKeyForUser(key.UserName, key.Name, id.Pretty()); err != nil {
		qm.l.Errorw(
			"failed to update database",
			"error", err.Error(),
			"user", key.UserName,
			"key_name", key.Name)
	} else {
		qm.l.Infow(
			"successfully processed key creation request",
			"user", key.UserName,
			"key_name", key.Name)
	}
	d.Ack(false)
	return // we must return here in order to trigger the wg.Done() defer
}
