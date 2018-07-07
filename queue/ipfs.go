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
func ProcessIpfsClusterQueue(msgs <-chan amqp.Delivery, db *gorm.DB) {
	var clusterPin IpfsClusterPin
	clusterManager := rtfs_cluster.Initialize()
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
}

// ProccessIPFSPins is used to process IPFS pin requests
func ProccessIPFSPins(msgs <-chan amqp.Delivery, db *gorm.DB) {
	userManager := models.NewUserManager(db)
	//uploadManager := models.NewUploadManager(db)
	networkManager := models.NewHostedIPFSNetworkManager(db)
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
				//TODO log 	and handle
				fmt.Println(err)
				fmt.Println("unauthorized access to private net ", pin.NetworkName)
				d.Ack(false)
				continue
			}
			url, err := networkManager.GetAPIURLByName(pin.NetworkName)
			if err != nil {
				//TODO log and handle
				fmt.Println(err)
				d.Ack(false)
			}
			apiURL = url
		}
		ipfsManager, err := rtfs.Initialize("", apiURL)
		if err != nil {
			//TODO log and handle
			// We aren't acknowledging this particular message
			// since it may be a temporary issue
			fmt.Println(err)
			continue
		}
		err = ipfsManager.Pin(pin.CID)
		if err != nil {
			//TODO log and handle
			// we aren't acknowlding this since it could be a temporary failure
			fmt.Println(err)
			fmt.Println("error pinning to network ", pin.NetworkName)
			continue
		}
	}
}

func ProccessIPFSFiles(msgs <-chan amqp.Delivery, cfg *config.TemporalConfig, db *gorm.DB) error {
	// construct the endpoint url to access our minio server
	endpoint := fmt.Sprintf("%s:%s", cfg.MINIO.Connection.IP, cfg.MINIO.Connection.Port)
	// grab our credentials for minio
	accessKey := cfg.MINIO.AccessKey
	secretKey := cfg.MINIO.SecretKey
	// setup our connection to local ipfs node
	ipfsManager, err := rtfs.Initialize("", "")
	if err != nil {
		return err
	}
	// setup our connection to minio
	minioManager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, false)
	if err != nil {
		return err
	}
	// process any received messages
	for d := range msgs {
		ipfsFile := IPFSFile{}
		// unmarshal the messagee
		err = json.Unmarshal(d.Body, &ipfsFile)
		if err != nil {
			//TODO: log and handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		// get object from minio
		obj, err := minioManager.GetObject(ipfsFile.BucketName, ipfsFile.ObjectName, minio.GetObjectOptions{})
		if err != nil {
			//TODO: log and handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		// add object to IPFs
		resp, err := ipfsManager.Shell.Add(obj)
		if err != nil {
			//TODO: log and handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
		err = minioManager.RemoveObject(ipfsFile.BucketName, ipfsFile.ObjectName)
		if err != nil {
			//TODO: log and handle
			fmt.Println(err)
			d.Ack(false)
			continue
		}
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

		}
		//TEMPORARY, will need to add logic here for processing of of records already in the database
		d.Ack(false)
	}
	return nil
}
