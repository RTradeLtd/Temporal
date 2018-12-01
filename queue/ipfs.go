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

	log "github.com/sirupsen/logrus"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/kaas"
	"github.com/RTradeLtd/rtfs"

	"github.com/RTradeLtd/config"

	"github.com/RTradeLtd/database/models"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"

	pb "github.com/RTradeLtd/grpc/krab"
	ci "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

// ProcessIPFSKeyCreation is used to create IPFS keys
func (qm *Manager) ProcessIPFSKeyCreation(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	kb, err := kaas.NewClient(cfg.Endpoints)
	if err != nil {
		return err
	}
	userManager := models.NewUserManager(db)
	qm.LogInfo("processing ipfs key creation requests")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go func(d amqp.Delivery) {
				defer wg.Done()
				qm.LogInfo("new message detected")
				key := IPFSKeyCreation{}
				err = json.Unmarshal(d.Body, &key)
				if err != nil {
					qm.LogError(err, "failed to unmarshal message")
					d.Ack(false)
					return
				}
				if key.NetworkName != "public" {
					qm.refundCredits(key.UserName, "key", key.CreditCost, db)
					qm.LogError(err, "private network key creation not yet supported")
					d.Ack(false)
					return
				}
				var (
					keyTypeInt int
					bitsInt    int
				)
				switch key.Type {
				case "rsa":
					keyTypeInt = ci.RSA
					if key.Size > 4096 || key.Size < 2048 {
						bitsInt = 2048
					} else {
						bitsInt = key.Size
					}
				case "ed25519":
					keyTypeInt = ci.Ed25519
					bitsInt = 256
				default:
					qm.refundCredits(key.UserName, "key", key.CreditCost, db)
					qm.LogError(err, "invalid key type must be ed25519 or rsa")
					d.Ack(false)
					return
				}
				keyName := fmt.Sprintf("%s-%s", key.UserName, key.Name)
				pk, _, err := ci.GenerateKeyPair(keyTypeInt, bitsInt)
				if err != nil {
					qm.refundCredits(key.UserName, "key", key.CreditCost, db)
					qm.LogError(err, "failed to create key")
					d.Ack(false)
					return
				}
				pkBytes, err := pk.Bytes()
				if err != nil {
					qm.refundCredits(key.UserName, "key", key.CreditCost, db)
					qm.LogError(err, "failed to create key")
					d.Ack(false)
					return
				}
				if _, err := kb.PutPrivateKey(context.Background(), &pb.KeyPut{Name: keyName, PrivateKey: pkBytes}); err != nil {
					qm.refundCredits(key.UserName, "key", key.CreditCost, db)
					qm.LogError(err, "failed to create key")
					d.Ack(false)
					return
				}
				// will need a refund, as this would mean that the key was improperly generated
				id, err := peer.IDFromPrivateKey(pk)
				if err != nil {
					qm.refundCredits(key.UserName, "key", key.CreditCost, db)
					qm.LogError(err, "failed to get id from private key")
					d.Ack(false)
					return
				}
				// doesn't need a refund, key wasgenerated, but information not saved to db
				if err := userManager.AddIPFSKeyForUser(key.UserName, keyName, id.Pretty()); err != nil {
					qm.LogError(err, "failed to add ipfs key to database")
					d.Ack(false)
					return
				}
				qm.LogInfo("successfully processed ipfs key creation request")
				d.Ack(false)
			}(d)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
		}
	}
}

// ProccessIPFSPins is used to process IPFS pin requests
func (qm *Manager) ProccessIPFSPins(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	userManager := models.NewUserManager(db)
	//uploadManager := models.NewUploadManager(db)
	networkManager := models.NewHostedIPFSNetworkManager(db)
	uploadManager := models.NewUploadManager(db)
	qmCluster, err := Initialize(IpfsClusterPinQueue, cfg.RabbitMQ.URL, true, false)
	if err != nil {
		qm.LogError(err, "failed to initialize cluster pin queue connection")
		return err
	}
	qm.LogInfo("processing ipfs pins")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go func(d amqp.Delivery) {
				defer wg.Done()
				qm.LogInfo("new message detected")
				pin := &IPFSPin{}
				err := json.Unmarshal(d.Body, pin)
				if err != nil {
					qm.LogError(err, "failed to unmarshal message")
					d.Ack(false)
					return
				}
				apiURL := ""
				if pin.NetworkName != "public" {
					canAccess, err := userManager.CheckIfUserHasAccessToNetwork(pin.UserName, pin.NetworkName)
					if err != nil {
						qm.refundCredits(pin.UserName, "pin", pin.CreditCost, db)
						qm.LogError(err, "failed to lookup private network in database")
						d.Ack(false)
						return
					}
					if !canAccess {
						qm.refundCredits(pin.UserName, "pin", pin.CreditCost, db)
						qm.LogError(errors.New("user does not have access to private network"), "invalid private network access")
						d.Ack(false)
						return
					}
					url, err := networkManager.GetAPIURLByName(pin.NetworkName)
					if err != nil {
						qm.refundCredits(pin.UserName, "pin", pin.CreditCost, db)
						qm.LogError(err, "failed to search for api url")
						d.Ack(false)
						return
					}
					apiURL = url
				}
				qm.LogInfo("initializing connection to ipfs")
				ipfsManager, err := rtfs.NewManager(apiURL, nil, time.Minute*10)
				if err != nil {
					qm.refundCredits(pin.UserName, "pin", pin.CreditCost, db)
					qm.LogError(err, "failed to initialize connection to ipfs")
					d.Ack(false)
					return
				}
				qm.LogInfo("pinning hash to ipfs")
				if err = ipfsManager.Pin(pin.CID); err != nil {
					qm.refundCredits(pin.UserName, "pin", pin.CreditCost, db)
					qm.LogError(err, "failed to pin hash to ipfs")
					d.Ack(false)
					return
				}
				qm.LogInfo("successfully pinned hash to ipfs")
				clusterAddMsg := IPFSClusterPin{
					CID:              pin.CID,
					NetworkName:      pin.NetworkName,
					HoldTimeInMonths: pin.HoldTimeInMonths,
					UserName:         pin.UserName,
				}
				// no need to do any payment processing on cluster as it has been handled already
				clusterAddMsg.CreditCost = 0
				if err = qmCluster.PublishMessage(clusterAddMsg); err != nil {
					qm.LogError(err, "failed to publish cluster pin message to rabbitmq")
					d.Ack(false)
					return
				}
				_, err = uploadManager.FindUploadByHashAndNetwork(pin.CID, pin.NetworkName)
				if err != nil && err != gorm.ErrRecordNotFound {
					qm.LogError(err, "failed to find upload in database")
					d.Ack(false)
					return
				}
				if err == gorm.ErrRecordNotFound {
					_, err = uploadManager.NewUpload(pin.CID, "pin", models.UploadOptions{
						NetworkName:      pin.NetworkName,
						Username:         pin.UserName,
						HoldTimeInMonths: pin.HoldTimeInMonths})
					if err != nil {
						qm.LogError(err, "failed to create upload in database")
						d.Ack(false)
						return
					}
				} else {
					// the record already exists so we will update
					_, err = uploadManager.UpdateUpload(pin.HoldTimeInMonths, pin.UserName, pin.CID, pin.NetworkName)
					if err != nil {
						qm.LogError(err, "failed to update upload in database")
						d.Ack(false)
						return
					}
				}
				qm.LogInfo("successfully processed pin request")
				d.Ack(false)
			}(d)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
		}
	}
}

// ProccessIPFSFiles is used to process messages sent to rabbitmq to upload files to IPFS.
// This function is invoked with the advanced method of file uploads, and is significantly more resilient than
// the simple file upload method.
func (qm *Manager) ProccessIPFSFiles(ctx context.Context, wg *sync.WaitGroup, msgs <-chan amqp.Delivery, cfg *config.TemporalConfig, db *gorm.DB) error {
	service := qm.Logger.WithFields(log.Fields{
		"service": qm.QueueName,
	})
	ipfsManager, err := rtfs.NewManager(cfg.IPFS.APIConnection.Host+":"+cfg.IPFS.APIConnection.Port, nil, time.Minute*10)
	if err != nil {
		service.
			WithField("error", err.Error()).
			Error("failed to initialize connection to ipfs")
		return err
	}
	qmPin, err := Initialize(IpfsPinQueue, cfg.RabbitMQ.URL, true, false)
	if err != nil {
		service.
			WithField("error", err.Error()).
			Error("failed to initialize pin queue connection")
		return err
	}
	qmMongo, err := Initialize(MongoUpdateQueue, cfg.RabbitMQ.URL, true, false)
	if err != nil {
		service.WithField("error", err.Error()).
			Error("failed to initialize mongodb update queue")
		return err
	}
	ue := models.NewEncryptedUploadManager(db)
	userManager := models.NewUserManager(db)
	networkManager := models.NewHostedIPFSNetworkManager(db)
	uploadManager := models.NewUploadManager(db)
	service.Info("processing ipfs files")
	for {
		select {
		case d := <-msgs:
			wg.Add(1)
			go func(d amqp.Delivery) {
				defer wg.Done()
				service.Info("new message detected")
				ipfsFile := IPFSFile{}
				// unmarshal the messagee
				err = json.Unmarshal(d.Body, &ipfsFile)
				if err != nil {
					service.
						WithField("error", err.Error()).
						Error("failed to unmarshal message")
					d.Ack(false)
					return
				}
				// now that we have the minio host which is storing this object, we can connect
				// construct the endpoint url to access our minio server
				endpoint := fmt.Sprintf("%s:%s", ipfsFile.MinioHostIP, cfg.MINIO.Connection.Port)

				// grab our credentials for minio
				accessKey := cfg.MINIO.AccessKey
				secretKey := cfg.MINIO.SecretKey
				// setup our connection to minio
				minioManager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, false)
				if err != nil {
					service.
						WithField("error", err.Error()).
						Error("failed to initialize connection to minio")
					d.Ack(false)
					return
				}
				fileContext := service.WithFields(log.Fields{
					"user":    ipfsFile.UserName,
					"network": ipfsFile.NetworkName,
				})

				if ipfsFile.NetworkName != "public" {
					canAccess, err := userManager.CheckIfUserHasAccessToNetwork(ipfsFile.UserName, ipfsFile.NetworkName)
					if err != nil {
						fileContext.
							WithField("error", err.Error()).
							Error("failed to check database for user network access")
						d.Ack(false)
						return
					}
					if !canAccess {
						qm.refundCredits(ipfsFile.UserName, "file", ipfsFile.CreditCost, db)
						fileContext.Error("unauthorized access to private network")
						d.Ack(false)
						return
					}
					apiURLName, err := networkManager.GetAPIURLByName(ipfsFile.NetworkName)
					if err != nil {
						fileContext.
							WithField("error", err.Error()).
							Error("failed to look for api url by name")
						d.Ack(false)
						return
					}
					apiURL := apiURLName
					fileContext.Info("initializing connection to private ipfs network")
					ipfsManager, err = rtfs.NewManager(apiURL, nil, time.Minute*10)
					if err != nil {
						qm.refundCredits(ipfsFile.UserName, "file", ipfsFile.CreditCost, db)
						fileContext.
							WithField("error", err.Error()).
							Error("failed to initialize connection to private ipfs network")
						d.Ack(false)
						return
					}
				}

				fileContext.Info("retrieving object from minio")

				obj, err := minioManager.GetObject(ipfsFile.ObjectName, mini.GetObjectOptions{
					Bucket: ipfsFile.BucketName,
				})
				if err != nil {
					fileContext.
						WithField("error", err.Error()).
						Info("failed to retrieve object from minio")
					d.Ack(false)
					return
				}

				fileContext.Info("successfully retrieved object from minio, adding file to ipfs")
				resp, err := ipfsManager.Add(obj)
				if err != nil {
					qm.refundCredits(ipfsFile.UserName, "file", ipfsFile.CreditCost, db)
					fileContext.
						WithField("error", err.Error()).
						Info("failed to add file to ipfs")
					d.Ack(false)
					return
				}

				fileContext.Info("file successfully added to IPFS, forwarding pin request")

				holdTimeInt, err := strconv.ParseInt(ipfsFile.HoldTimeInMonths, 10, 64)
				if err != nil {
					fileContext.
						WithField("error", err.Error()).
						Warn("failed to parse string to int, using default of 1 month")
					holdTimeInt = 1
				}

				// we don't need to do any credit handling, as it has been done already
				pin := IPFSPin{
					CID:              resp,
					NetworkName:      ipfsFile.NetworkName,
					UserName:         ipfsFile.UserName,
					HoldTimeInMonths: holdTimeInt,
					CreditCost:       0,
				}

				err = qmPin.PublishMessageWithExchange(pin, PinExchange)
				if err != nil {
					fileContext.
						WithField("error", err.Error()).
						Warn("failed to publish message to pin queue")

				}

				_, err = uploadManager.FindUploadByHashAndNetwork(resp, ipfsFile.NetworkName)
				if err != nil && err != gorm.ErrRecordNotFound {
					fileContext.
						WithField("error", err.Error()).
						Error("failed to look for upload in database")
					d.Ack(false)
					return
				}
				if err == gorm.ErrRecordNotFound {
					if _, err = uploadManager.NewUpload(resp, "file", models.UploadOptions{
						NetworkName:      ipfsFile.NetworkName,
						Username:         ipfsFile.UserName,
						HoldTimeInMonths: holdTimeInt,
						Encrypted:        ipfsFile.Encrypted,
					}); err != nil {
						fileContext.
							WithField("error", err.Error()).
							Error("failed to create new upload in database")
						d.Ack(false)
						return
					}
				} else {
					_, err = uploadManager.UpdateUpload(holdTimeInt, ipfsFile.UserName, resp, ipfsFile.NetworkName)
					if err != nil {
						fileContext.
							WithField("error", err.Error()).
							Error("failed to update upload in database")
						d.Ack(false)
						return
					}
				}
				// if encrypted upload, do some special processing
				if ipfsFile.Encrypted {
					if _, err = ue.NewUpload(
						ipfsFile.UserName,
						ipfsFile.FileName,
						ipfsFile.NetworkName,
						resp,
					); err != nil {
						// we won't ack this, since we have already processed the upload and this is "extra processing"
						fileContext.
							WithField("error", err.Error()).
							Error("failed to upload database with encrypted upload")
					}
					sizeString := strconv.FormatInt(ipfsFile.FileSize, 10)
					mongoUpdate := MongoUpdate{
						DatabaseName:   cfg.Endpoints.MongoDB.DB,
						CollectionName: cfg.Endpoints.MongoDB.UploadCollection,
						Fields: map[string]string{
							"size":          sizeString,
							"user":          ipfsFile.UserName,
							"fileName":      ipfsFile.FileName,
							"fileNameLower": strings.ToLower(ipfsFile.FileName),
							"hash":          resp,
							"encrypted":     "true",
						},
					}
					if err = qmMongo.PublishMessage(mongoUpdate); err != nil {
						fileContext.
							WithField("error", err.Error()).
							Error("failed to publish message to mongo update queue")
					}
				}
				fileContext.Info("removing object from minio")
				err = minioManager.RemoveObject(ipfsFile.BucketName, ipfsFile.ObjectName)
				if err != nil {
					fileContext.
						WithField("error", err.Error()).
						Error("failed to remove object from minio")
					d.Ack(false)
					return
				}

				fileContext.Info("object removed from minio, successfully added to ipfs")
				d.Ack(false)
			}(d)
		case <-ctx.Done():
			qm.Close()
			wg.Done()
		}
	}
}
