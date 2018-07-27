package queue

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/minio/minio-go"

	"github.com/RTradeLtd/Temporal/mini"

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
	uploadManager := models.NewUploadManager(db)
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
			continue
		}
		apiURL := ""
		if pin.NetworkName != "public" {
			canAccess, err := userManager.CheckIfUserHasAccessToNetwork(pin.EthAddress, pin.NetworkName)
			if err != nil {
				//TODO log and handle
				fmt.Println("error checking for private network access", err)
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
				fmt.Println("error publishing message ", err)
			}
			//TODO log and handle
			// we aren't acknowlding this since it could be a temporary failure
			fmt.Println(err)
			fmt.Println("error pinning to network ", pin.NetworkName)
			continue
		}
		_, err = uploadManager.FindUploadByHashAndNetwork(pin.CID, pin.NetworkName)
		if err != nil && err != gorm.ErrRecordNotFound {
			fmt.Println("error getting model from database ", err)
			// decide what to do here
			d.Ack(false)
			continue
		}
		if err == gorm.ErrRecordNotFound {
			_, check := uploadManager.NewUpload(pin.CID, "pin", pin.NetworkName, pin.EthAddress, pin.HoldTimeInMonths)
			if check != nil {
				fmt.Println("error creating new upload ", check)
				// decide what to do ehre, who we should email, etcc...
				d.Ack(false)
				continue
			}
		}
		// the record already exists so we will update
		_, err = uploadManager.UpdateUpload(pin.HoldTimeInMonths, pin.EthAddress, pin.CID, pin.NetworkName)
		if err != nil {
			fmt.Println("error updating model in database ", err)
			// TODO: decide what to do, who we should email, etcc
			d.Ack(false)
			continue
		}
		d.Ack(false)
	}
	return nil
}

// ProccessIPFSFiles is used to process messages sent to rabbitmq to upload files to IPFS.
// This function is invoked with the advanced method of file uploads, and is significantly more resilient than
// the simple file upload method.
func ProccessIPFSFiles(msgs <-chan amqp.Delivery, cfg *config.TemporalConfig, db *gorm.DB) error {
	// construct the endpoint url to access our minio server
	endpoint := fmt.Sprintf("%s:%s", cfg.MINIO.Connection.IP, cfg.MINIO.Connection.Port)
	// grab our credentials for minio
	accessKey := cfg.MINIO.AccessKey
	secretKey := cfg.MINIO.SecretKey
	fmt.Println("setting up ipfs connection")
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
	qmFile, err := Initialize(IpfsFileQueue, cfg.RabbitMQ.URL)
	if err != nil {
		return err
	}
	qmEmail, err := Initialize(EmailSendQueue, cfg.RabbitMQ.URL)
	if err != nil {
		return err
	}
	userManager := models.NewUserManager(db)
	networkManager := models.NewHostedIPFSNetworkManager(db)
	uploadManager := models.NewUploadManager(db)
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
		fmt.Println("determining network")
		apiURL := ""
		// determing private network access rights
		if ipfsFile.NetworkName != "public" {
			canAccess, err := userManager.CheckIfUserHasAccessToNetwork(ipfsFile.EthAddress, ipfsFile.NetworkName)
			if err != nil {
				//TODO log and handle, decide how we would do this
				fmt.Println("error checking for private network access", err)
				d.Ack(false)
				continue
			}
			if !canAccess {
				addresses := []string{}
				addresses = append(addresses, ipfsFile.EthAddress)
				es := EmailSend{
					Subject:      IpfsPrivateNetworkUnauthorizedSubject,
					Content:      fmt.Sprintf("Unauthorized access to IPFS private network %s", ipfsFile.NetworkName),
					ContentType:  "",
					EthAddresses: addresses,
				}
				err = qmEmail.PublishMessage(es)
				if err != nil {
					//TODO log and handle
					fmt.Println(err)
				}
				//TODO log 	and handle
				fmt.Println("unauthorized access to private net ", ipfsFile.NetworkName)
				d.Ack(false)
				continue
			}
			apiURLName, err := networkManager.GetAPIURLByName(ipfsFile.NetworkName)
			if err != nil {
				//TODO send email, log, handle
				fmt.Println("error getting API url by name ", err)
				d.Ack(false)
				continue
			}
			apiURL = apiURLName
			ipfsManager, err = rtfs.Initialize("", apiURL)
			if err != nil {
				addresses := []string{}
				addresses = append(addresses, ipfsFile.EthAddress)
				es := EmailSend{
					Subject:      IpfsInitializationFailedSubject,
					Content:      fmt.Sprintf("Connection to IPFS failed due to the following error %s", err),
					ContentType:  "",
					EthAddresses: addresses,
				}
				errOne := qmEmail.PublishMessage(es)
				if errOne != nil {
					fmt.Println("error publishing message ", err)
				}
				fmt.Println(err)
				d.Ack(false)
				continue
			}
		}

		fmt.Println("retrieving file from minio")
		// get object from minio
		obj, err := minioManager.GetObject(ipfsFile.BucketName, ipfsFile.ObjectName, minio.GetObjectOptions{})
		if err != nil {
			//TODO: log and handle, should we email them when this fails?
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		fmt.Println("file retrieved from minio")
		// add object to IPFs
		fmt.Println("adding file to ipfs")
		resp, err := ipfsManager.Shell.Add(obj)
		if err != nil {
			//TODO: decide how to handle email failures
			addresses := []string{}
			addresses = append(addresses, ipfsFile.EthAddress)
			es := EmailSend{
				Subject:      IpfsFileFailedSubject,
				Content:      fmt.Sprintf(IpfsFileFailedContent, ipfsFile.ObjectName, ipfsFile.NetworkName),
				ContentType:  "",
				EthAddresses: addresses,
			}
			errOne := qmEmail.PublishMessage(es)
			if errOne != nil {
				fmt.Println(errOne)
			}
			//TODO: log and handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		holdTimeInt, err := strconv.ParseInt(ipfsFile.HoldTimeInMonths, 10, 64)
		if err != nil {
			fmt.Println("erorr parsing string to int ", err)
			//TODO decide how to handle, etc..
			d.Ack(false)
			continue
		}
		ipfsPin := IPFSPin{
			CID:              resp,
			NetworkName:      ipfsFile.NetworkName,
			EthAddress:       ipfsFile.EthAddress,
			HoldTimeInMonths: holdTimeInt,
		}
		err = qmFile.PublishMessageWithExchange(ipfsPin, PinExchange)
		if err != nil {
			// this we will won't ack, or continue on since the file has already been added to ipfs and can be pinned seperately
			fmt.Println("error publishing ipfs pin message to the pin exchange ", err)
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
		// TODO: decide whether or not we should email on "backend" failures
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
		//THIS PART IS EXTREMELY INEFFICIENT AND NEEDS TO BE RE-EXAMINED
		// if we have an error of type record not found, lets build a fresh model to save in the database
		if check.Error == gorm.ErrRecordNotFound {
			_, err = uploadManager.NewUpload(resp, "file", ipfsFile.NetworkName, ipfsFile.EthAddress, holdTimeInt)
			if err != nil {
				//TODO decide how we should handle this
				fmt.Println("error creating new upload in database ", err)
				d.Ack(false)
				continue
			}
			d.Ack(false)
			continue
		}
		_, err = uploadManager.UpdateUpload(holdTimeInt, ipfsFile.EthAddress, resp, ipfsFile.NetworkName)
		if err != nil {
			//TODO decide how to handle
			fmt.Println("error updating upload in database ", err)
			d.Ack(false)
			continue
		}
		//TEMPORARY, will need to add logic here for processing of of records already in the database
		d.Ack(false)
	}
	return nil
}
