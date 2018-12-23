package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/RTradeLtd/Temporal/log"
	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/kaas"
	"github.com/RTradeLtd/rtfs"

	"github.com/RTradeLtd/database/models"
	"github.com/jinzhu/gorm"
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
	qm.l.Info("processing ipfs pins")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processIPFSPin(d, wg, userManager, networkManager, uploadManager, qmCluster)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
			return nil
		}
	}
}

// ProccessIPFSFiles is used to process messages sent to rabbitmq to upload files to IPFS.
// This queue is invoked by advanced file upload requests
func (qm *Manager) ProccessIPFSFiles(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery) error {
	ipfsManager, err := rtfs.NewManager(qm.cfg.IPFS.APIConnection.Host+":"+qm.cfg.IPFS.APIConnection.Port, nil, time.Minute*10)
	if err != nil {
		qm.l.Errorw("failed to initialize connection to ipfs", "error", err.Error())
		return err
	}
	logger, err := log.NewLogger(qm.cfg.LogDir+"pin_publisher.log", false)
	if err != nil {
		return err
	}
	// initialize connection to pin queues
	qmPin, err := New(IpfsPinQueue, qm.cfg.RabbitMQ.URL, true, logger)
	if err != nil {
		qm.l.Errorw("failed to initialize pin queue connection", "error", err.Error())
		return err
	}
	ue := models.NewEncryptedUploadManager(qm.db)
	userManager := models.NewUserManager(qm.db)
	networkManager := models.NewHostedIPFSNetworkManager(qm.db)
	qm.l.Info("processing ipfs files")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go qm.processIPFSFile(d, wg, ue, userManager, networkManager, ipfsManager, qmPin)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
			return nil
		}
	}
}

func (qm *Manager) processIPFSPin(d amqp.Delivery, wg *sync.WaitGroup, usrm *models.UserManager, nm *models.IPFSNetworkManager, upldm *models.UploadManager, qmCluster *Manager) {
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
			qm.refundCredits(pin.UserName, "pin", pin.CreditCost)
			qm.l.Errorw("failed to lookup private network in database", "error", err.Error())
			d.Ack(false)
			return
		}
		if !canAccess {
			qm.refundCredits(pin.UserName, "pin", pin.CreditCost)
			qm.l.Errorw(
				"unauthorized private network access",
				"error", errors.New("user does not have access to private network").Error(),
				"user", pin.UserName)
			d.Ack(false)
			return
		}
		apiURL, err = nm.GetAPIURLByName(pin.NetworkName)
		if err != nil {
			qm.refundCredits(pin.UserName, "pin", pin.CreditCost)
			qm.l.Errorw(
				"failed to search for api url",
				"error", err.Error(),
				"user", pin.UserName)
			d.Ack(false)
			return
		}
	}
	qm.l.Infow(
		"initializing connection to ipfs",
		"user", pin.UserName)
	// connect to ipfs
	ipfsManager, err := rtfs.NewManager(apiURL, nil, time.Minute*10)
	if err != nil {
		qm.refundCredits(pin.UserName, "pin", pin.CreditCost)
		qm.l.Infow(
			"failed to initialize connection to ipfs",
			"error", err.Error(),
			"user", pin.UserName,
			"network", pin.NetworkName)
		d.Ack(false)
		return
	}
	qm.l.Infow(
		"pinning hash to ipfs",
		"cid", pin.CID,
		"user", pin.UserName,
		"network", pin.NetworkName)
	// pin the content
	if err = ipfsManager.Pin(pin.CID); err != nil {
		qm.refundCredits(pin.UserName, "pin", pin.CreditCost)
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
	if pin.NetworkName != "public" {
		// private networks do not yet support cluster so skip additional processing
		d.Ack(false)
		return
	}
	clusterAddMsg := IPFSClusterPin{
		CID:              pin.CID,
		NetworkName:      pin.NetworkName,
		HoldTimeInMonths: pin.HoldTimeInMonths,
		UserName:         pin.UserName,
	}
	// do not perform credit handling, as the content is already pinned
	clusterAddMsg.CreditCost = 0
	if err = qmCluster.PublishMessage(clusterAddMsg); err != nil {
		qm.l.Errorw(
			"failed to publish cluster pin message to rabbitmq",
			"error", err.Error(),
			"user", pin.UserName)
		d.Ack(false)
		return
	}
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

func (qm *Manager) processIPFSFile(d amqp.Delivery, wg *sync.WaitGroup, ue *models.EncryptedUploadManager, um *models.UserManager, nm *models.IPFSNetworkManager, ipfs *rtfs.IpfsManager, qmPin *Manager) {
	defer wg.Done()
	qm.l.Info("new file upload request detected")
	ipfsFile := IPFSFile{}
	// unmarshal the messagee
	if err := json.Unmarshal(d.Body, &ipfsFile); err != nil {
		qm.l.Errorw(
			"failed to unmarshal message",
			"error", err.Error())
		d.Ack(false)
		return
	}
	// get the connection url for the minio host temporarily storing this file
	endpoint := fmt.Sprintf("%s:%s", ipfsFile.MinioHostIP, qm.cfg.MINIO.Connection.Port)
	// grab our credentials for minio
	accessKey := qm.cfg.MINIO.AccessKey
	secretKey := qm.cfg.MINIO.SecretKey
	// setup our connection to minio
	minioManager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, false)
	if err != nil {
		qm.l.Errorw(
			"failed to connect to minio",
			"error", err.Error(),
			"user", ipfsFile.UserName)
		qm.refundCredits(ipfsFile.UserName, "file", ipfsFile.CreditCost)
		d.Ack(false)
		return
	}
	// determine whether or not the request is for a private network
	if ipfsFile.NetworkName != "public" {
		canAccess, err := um.CheckIfUserHasAccessToNetwork(ipfsFile.UserName, ipfsFile.NetworkName)
		if err != nil {
			qm.l.Errorw(
				"failed to check database for private network access",
				"error", err.Error(),
				"user", ipfsFile.UserName,
				"network", ipfsFile.NetworkName)
			qm.refundCredits(ipfsFile.UserName, "file", ipfsFile.CreditCost)
			d.Ack(false)
			return
		}
		if !canAccess {
			qm.l.Errorw(
				"unauthorized private network access",
				"error", errors.New("user does not have access to private network").Error(),
				"user", ipfsFile.UserName,
				"network", ipfsFile.NetworkName)
			qm.refundCredits(ipfsFile.UserName, "file", ipfsFile.CreditCost)
			d.Ack(false)
			return
		}
		apiURLName, err := nm.GetAPIURLByName(ipfsFile.NetworkName)
		if err != nil {
			qm.l.Errorw(
				"failed to search for api url",
				"error", err.Error(),
				"user", ipfsFile.UserName)
			qm.refundCredits(ipfsFile.UserName, "file", ipfsFile.CreditCost)
			d.Ack(false)
			return
		}
		ipfs, err = rtfs.NewManager(apiURLName, nil, time.Minute*10)
		if err != nil {
			qm.l.Errorw(
				"failed to initialize connection to private ipfs network",
				"error", err.Error(),
				"user", ipfsFile.UserName,
				"network", ipfsFile.NetworkName)
			qm.refundCredits(ipfsFile.UserName, "file", ipfsFile.CreditCost)
			d.Ack(false)
			return
		}
	}
	// start upload processing
	// 1. retrieve object from minio
	// 2. add file to ipfs
	// 3. send a pin request
	// 4. *optional* perform as needed special processing
	// NOTE: we do not trigger a database update here, as that is handled by the pin queue
	qm.l.Infow(
		"retrieving object from minio",
		"object", ipfsFile.ObjectName,
		"user", ipfsFile.UserName,
		"network", ipfsFile.NetworkName)
	obj, err := minioManager.GetObject(ipfsFile.ObjectName, mini.GetObjectOptions{
		Bucket: ipfsFile.BucketName,
	})
	if err != nil {
		qm.l.Errorw(
			"failed to retrieve object from minio",
			"error", err.Error(),
			"object", ipfsFile.ObjectName,
			"user", ipfsFile.UserName,
			"network", ipfsFile.NetworkName)
		qm.refundCredits(ipfsFile.UserName, "file", ipfsFile.CreditCost)
		d.Ack(false)
		return
	}
	qm.l.Infow(
		"uploading file to ipfs",
		"object", ipfsFile.ObjectName,
		"user", ipfsFile.UserName,
		"network", ipfsFile.NetworkName)
	resp, err := ipfs.Add(obj)
	if err != nil {
		qm.l.Errorw(
			"failed to upload file to ipfs",
			"error", err.Error(),
			"object", ipfsFile.ObjectName,
			"user", ipfsFile.UserName,
			"network", ipfsFile.NetworkName)
		qm.refundCredits(ipfsFile.UserName, "file", ipfsFile.CreditCost)
		d.Ack(false)
		return
	}
	qm.l.Infow(
		"successfully uploaded file to ipfs",
		"object", ipfsFile.ObjectName,
		"cid", resp,
		"user", ipfsFile.UserName,
		"network", ipfsFile.NetworkName)
	holdTimeInt, err := strconv.ParseInt(ipfsFile.HoldTimeInMonths, 10, 64)
	if err != nil {
		holdTimeInt = 1
	}
	pin := IPFSPin{
		CID:              resp,
		NetworkName:      ipfsFile.NetworkName,
		UserName:         ipfsFile.UserName,
		HoldTimeInMonths: holdTimeInt,
		CreditCost:       0,
	}
	// if encrypted upload, do some special processing
	if ipfsFile.Encrypted {
		if _, err = ue.NewUpload(
			ipfsFile.UserName,
			ipfsFile.FileName,
			ipfsFile.NetworkName,
			resp,
		); err != nil {
			qm.l.Errorw(
				"failed to update database with encrypted upload",
				"error", err.Error(),
				"cid", resp,
				"user", ipfsFile.UserName,
				"network", ipfsFile.NetworkName)
			// dont ack/fail yet as there is still additional processing to do
		}
	}
	if err = qmPin.PublishMessageWithExchange(pin, PinExchange); err != nil {
		qm.l.Errorw(
			"failed to publish pin request to rabbitmq",
			"error", err.Error(),
			"cid", resp,
			"user", ipfsFile.UserName,
			"network", ipfsFile.NetworkName)
		// we want to ack and return here, as removing the file from minio
		// could lead to us being unable to reprocess this later on
		d.Ack(false)
		return
	}
	qm.l.Infow(
		"removing object from minio",
		"object", ipfsFile.ObjectName,
		"user", ipfsFile.UserName)
	if err = minioManager.RemoveObject(ipfsFile.BucketName, ipfsFile.ObjectName); err != nil {
		qm.l.Errorw(
			"failed to remove object from minio",
			"error", err.Error(),
			"object", ipfsFile.ObjectName,
			"user", ipfsFile.UserName)
	} else {
		qm.l.Infow(
			"removed object from minio",
			"object", ipfsFile.ObjectName,
			"user", ipfsFile.UserName)
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
		qm.refundCredits(key.UserName, "key", key.CreditCost)
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
		qm.refundCredits(key.UserName, "key", key.CreditCost)
		d.Ack(false)
		return
	}
	// generate the appropriate keypair
	pk, _, err := ci.GenerateKeyPair(keyTypeInt, bitsInt)
	if err != nil {
		qm.refundCredits(key.UserName, "key", key.CreditCost)
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
		qm.refundCredits(key.UserName, "key", key.CreditCost)
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
		qm.refundCredits(key.UserName, "key", key.CreditCost)
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
		qm.refundCredits(key.UserName, "key", key.CreditCost)
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
