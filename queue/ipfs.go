package queue

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/minio/minio-go"

	"github.com/RTradeLtd/Temporal/mini"
	"github.com/RTradeLtd/Temporal/utils"

	"github.com/RTradeLtd/Temporal/config"
	"github.com/RTradeLtd/Temporal/rtfs"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/rtfs_cluster"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// ProcessIpfsClusterQueue is used to process msgs sent to the ipfs cluster queue
func ProcessIpfsClusterQueue(msgs <-chan amqp.Delivery, db *gorm.DB) error {
	var clusterPin IpfsClusterPin
	clusterManager, err := rtfs_cluster.Initialize()
	if err != nil {
		return err
	}
	for d := range msgs {
		err := json.Unmarshal(d.Body, &clusterPin)
		if err != nil {
			fmt.Println("error unmarshaling data ", err)
			// TODO: handle error
			d.Ack(false)
			continue
		}
		contentHash := clusterPin.CID
		decodedContentHash, err := clusterManager.DecodeHashString(contentHash)
		if err != nil {
			fmt.Println("error decoded content hash to cid ", err)
			//TODO: handle error
			d.Ack(false)
			continue
		}
		err = clusterManager.Pin(decodedContentHash)
		if err != nil {
			fmt.Println("error pinning to cluster ", err)
			//TODO: handle error
			d.Ack(false)
			continue
		}
		fmt.Println("content pinned to cluster ", contentHash)
		d.Ack(false)

	}
	return nil
}

// ProccessIPFSPins is used to process IPFS pin requests
func ProccessIPFSPins(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	userManager := models.NewUserManager(db)
	//uploadManager := models.NewUploadManager(db)
	networkManager := models.NewHostedIPFSNetworkManager(db)
	qm, err := Initialize(EmailSendQueue, cfg.RabbitMQ.URL)
	if err != nil {
		return err
	}
	for d := range msgs {
		pin := &IPFSPin{}
		err := json.Unmarshal(d.Body, pin)
		if err != nil {
			//TODO log and handle
			fmt.Println(err)
			d.Ack(false)
		}
		apiURL := ""
		if pin.NetworkName != "public" {
			canAccess, err := userManager.CheckIfUserHasAccessToNetwork(pin.EthAddress, pin.NetworkName)
			if err != nil {
				//TODO log and handle
				fmt.Println(err)
				d.Ack(false)
				continue
			}
			if !canAccess {
				addresses := []string{}
				addresses = append(addresses, pin.EthAddress)
				es := EmailSend{
					Subject:      IpfsPrivateNetworkUnauthorizedSubject,
					Content:      fmt.Sprintf("Unauthorized access to IPFS private network %s", pin.NetworkName),
					ContentType:  "",
					EthAddresses: addresses,
				}
				err = qm.PublishMessage(es)
				if err != nil {
					//TODO log and handle
					fmt.Println(err)
				}
				//TODO log 	and handle
				fmt.Println("unauthorized access to private net ", pin.NetworkName)
				d.Ack(false)
				continue
			}
			url, err := networkManager.GetAPIURLByName(pin.NetworkName)
			if err != nil {
				//TODO: decide if we should send out an email
				fmt.Println(err)
				d.Ack(false)
				continue
			}
			apiURL = url
		}
		ipfsManager, err := rtfs.Initialize("", apiURL)
		if err != nil {
			addresses := []string{}
			addresses = append(addresses, pin.EthAddress)
			es := EmailSend{
				Subject:      IpfsInitializationFailedSubject,
				Content:      fmt.Sprintf("Connection to IPFS failed due to the following error %s", err),
				ContentType:  "",
				EthAddresses: addresses,
			}
			errOne := qm.PublishMessage(es)
			if errOne != nil {
				// For this, we will not ack since we want to be able to send messages
				fmt.Println("error publishing message ", err)
				continue
			}
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		err = ipfsManager.Pin(pin.CID)
		if err != nil {
			addresses := []string{}
			addresses = append(addresses, pin.EthAddress)
			es := EmailSend{
				Subject:      IpfsPinFailedSubject,
				Content:      fmt.Sprintf(IpfsPinFailedContent, pin.CID, pin.NetworkName, err),
				ContentType:  "",
				EthAddresses: addresses,
			}
			errOne := qm.PublishMessage(es)
			if errOne != nil {
				// For now we aren't acking this, however this needs to be
				// reevluated later on
				fmt.Println("error publishing message ", err)
				continue
			}
			//TODO log and handle
			// we aren't acknowlding this since it could be a temporary failure
			fmt.Println(err)
			fmt.Println("error pinning to network ", pin.NetworkName)
			continue
		}
	}
	return nil
}

func ProccessIPFSFiles(msgs <-chan amqp.Delivery, cfg *config.TemporalConfig, db *gorm.DB) error {
	// construct the endpoint url to access our minio server
	endpoint := fmt.Sprintf("%s:%s", cfg.MINIO.Connection.IP, cfg.MINIO.Connection.Port)
	// grab our credentials for minio
	accessKey := cfg.MINIO.AccessKey
	secretKey := cfg.MINIO.SecretKey
	fmt.Println("setting up ipfs connection")
	// setup our connection to local ipfs node
	ipfsManager, err := rtfs.Initialize("", "")
	if err != nil {
		return err
	}
	fmt.Println("ipfs connection setup")
	fmt.Println("setting up minio connection")
	// setup our connection to minio
	minioManager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, false)
	if err != nil {
		return err
	}
	fmt.Println("minio connection setup")
	// process any received messages
	fmt.Println("processing ipfs file messages")
	for d := range msgs {
		fmt.Println("file received")
		ipfsFile := IPFSFile{}
		// unmarshal the messagee
		err = json.Unmarshal(d.Body, &ipfsFile)
		if err != nil {
			//TODO: log and handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		fmt.Println("retrieving file from minio")
		// get object from minio
		obj, err := minioManager.GetObject(ipfsFile.BucketName, ipfsFile.ObjectName, minio.GetObjectOptions{})
		if err != nil {
			//TODO: log and handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		fmt.Println("file retrieved from minio")
		// add object to IPFs
		fmt.Println("adding file to ipfs")
		resp, err := ipfsManager.Shell.Add(obj)
		if err != nil {
			//TODO: log and handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		fmt.Println("file added to ipfs")
		fmt.Println("removing object from minio")
		err = minioManager.RemoveObject(ipfsFile.BucketName, ipfsFile.ObjectName)
		if err != nil {
			//TODO: log and handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		fmt.Println("object removed from minio")
		upload := models.Upload{}
		// find a model from the database matching the content hash and network name
		check := db.Where("hash = ? AND network_name = ?", resp, ipfsFile.NetworkName).First(&upload)
		// if we have an error, that is not of type record not found fail temporarily
		if check.Error != nil && check.Error != gorm.ErrRecordNotFound {
			//TODO: log and handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		// if we have an error of type record not found, lets build a fresh model to save in the database
		if check.Error == gorm.ErrRecordNotFound {
			upload.Hash = resp
			holdInt, err := strconv.Atoi(ipfsFile.HoldTimeInMonths)
			holdInt64, err := strconv.ParseInt(ipfsFile.HoldTimeInMonths, 10, 64)
			if err != nil {
				//TODO: log and handle
				fmt.Println(err)
				d.Ack(false)
				continue
			}
			if err != nil {
				//TODO: log and handle
				fmt.Println(err)
				d.Ack(false)
				continue
			}
			gcd := utils.CalculateGarbageCollectDate(holdInt)
			upload.GarbageCollectDate = gcd
			upload.HoldTimeInMonths = holdInt64
			upload.NetworkName = ipfsFile.NetworkName
			upload.UploadAddress = ipfsFile.EthAddress
			upload.UploaderAddresses = append(upload.UploaderAddresses, ipfsFile.EthAddress)
			if chk := db.Create(&upload); chk.Error != nil {
				//TODO: log and handle
				fmt.Println(err)
				d.Ack(false)
				continue
			}
			fmt.Println("file added to database")
			d.Ack(false)
			continue
		}
		//TEMPORARY, will need to add logic here for processing of of records already in the database
		d.Ack(false)
	}
	return nil
}
