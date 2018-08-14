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
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"

	ci "github.com/libp2p/go-libp2p-crypto"
	peer "github.com/libp2p/go-libp2p-peer"
)

// ProcessIPFSKeyCreation is used to create IPFS keys
func ProcessIPFSKeyCreation(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	manager, err := rtfs.Initialize("", "")
	if err != nil {
		return err
	}
	// load the keystore manager
	err = manager.CreateKeystoreManager()
	if err != nil {
		return err
	}
	userManager := models.NewUserManager(db)
	fmt.Println("processing ipfs key creation")
	for d := range msgs {
		key := IPFSKeyCreation{}
		err = json.Unmarshal(d.Body, &key)
		if err != nil {
			fmt.Println("error unmarshaling message ", err)
			d.Ack(false)
			continue
		}
		var keyTypeInt int
		var bitsInt int
		switch key.Type {
		case "rsa":
			keyTypeInt = ci.RSA
			if key.Size > 4096 {
				fmt.Println("key size generation greater than 4096 not supported")
				d.Ack(false)
				continue
			}
			bitsInt = key.Size
		case "ed25519":
			keyTypeInt = ci.Ed25519
			bitsInt = 256
		default:
			fmt.Println("unsupported key type")
		}
		pk, err := manager.KeystoreManager.CreateAndSaveKey(key.Name, keyTypeInt, bitsInt)
		if err != nil {
			fmt.Println("error creating key ", err)
			d.Ack(false)
			continue
		}

		id, err := peer.IDFromPrivateKey(pk)
		if err != nil {
			fmt.Println("failed to get id from private key ", err)
			d.Ack(false)
			continue
		}
		err = userManager.AddIPFSKeyForUser(key.EthAddress, key.Name, id.Pretty())
		if err != nil {
			fmt.Println("error adding ipfs key for user to database ", err)
			d.Ack(false)
			continue
		}
		fmt.Println("successfully created and saved key")
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
	qmCluster, err := Initialize(IpfsClusterAddQueue, cfg.RabbitMQ.URL)
	if err != nil {
		return err
	}
	for d := range msgs {
		fmt.Println("detected new content")
		pin := &IPFSPin{}
		err := json.Unmarshal(d.Body, pin)
		if err != nil {
			fmt.Println("failed to unmarshal response", err)
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
				fmt.Println("error getting api url by name ", err)
				d.Ack(false)
				continue
			}
			apiURL = url
		}
		fmt.Println("initializing ipfs")
		ipfsManager, err := rtfs.Initialize("", apiURL)
		if err != nil {
			fmt.Println("error initializing IPFS", err)
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
				fmt.Println("error publishing message to email queue", errOne)
			}
			fmt.Println("error initializing ipfs", err)
			d.Ack(false)
			continue
		}
		fmt.Printf("pinning content hash %s to ipfs\n", pin.CID)
		err = ipfsManager.Pin(pin.CID)
		if err != nil {
			fmt.Println("error pinning content to ipfs")
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
			d.Ack(false)
			continue
		}
		fmt.Println("successfully pinned content to ipfs")
		// automatically trigger a cluster add0
		go func() {
			clusterAddMsg := IPFSClusterAdd{
				CID:         pin.CID,
				NetworkName: pin.NetworkName,
			}
			qmCluster.PublishMessage(clusterAddMsg)
		}()
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
			d.Ack(false)
			continue
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

// ProcessIPFSPinRemovals is used to listen for and process any IPFS pin removals.
// This queue must be running on each of the IPFS nodes, and we must eventually run checks
// to ensure that pins were actually removed
func ProcessIPFSPinRemovals(msgs <-chan amqp.Delivery, cfg *config.TemporalConfig, db *gorm.DB) error {
	userManager := models.NewUserManager(db)
	networkManager := models.NewHostedIPFSNetworkManager(db)
	qmEmail, err := Initialize(EmailSendQueue, cfg.RabbitMQ.URL)
	if err != nil {
		return err
	}
	for d := range msgs {
		rm := IPFSPinRemoval{}
		err := json.Unmarshal(d.Body, &rm)
		if err != nil {
			//TODO: log and handle
			fmt.Println("error unmarshaling ", err)
			d.Ack(false)
			continue
		}
		apiURL := ""
		if rm.NetworkName != "public" {
			canAccess, err := userManager.CheckIfUserHasAccessToNetwork(rm.EthAddress, rm.NetworkName)
			if err != nil {
				//TODO: log and handle
				fmt.Println("error checking for network access ", err)
				d.Ack(false)
				continue
			}
			if !canAccess {
				addresses := []string{}
				addresses = append(addresses, rm.EthAddress)
				es := EmailSend{
					Subject:      IpfsPrivateNetworkUnauthorizedSubject,
					Content:      fmt.Sprintf("Unauthorized access to IPFS private network %s", rm.NetworkName),
					ContentType:  "",
					EthAddresses: addresses,
				}
				err = qmEmail.PublishMessage(es)
				if err != nil {
					//TODO log and handle
					fmt.Println(err)
				}
				//TODO log 	and handle
				fmt.Println("unauthorized access to private net ", rm.NetworkName)
				d.Ack(false)
				continue
			}
			apiURL, err = networkManager.GetAPIURLByName(rm.NetworkName)
			if err != nil {
				//TODO log and handle
				fmt.Println("failed to get api url for private network ", err)
				d.Ack(false)
				continue
			}
		}
		ipfsManager, err := rtfs.Initialize("", apiURL)
		if err != nil {
			addresses := []string{rm.EthAddress}
			es := EmailSend{
				Subject:      IpfsInitializationFailedSubject,
				Content:      fmt.Sprintf("Failed to connect to IPFS network %s for reason %s", rm.NetworkName, err),
				ContentType:  "",
				EthAddresses: addresses,
			}
			errOne := qmEmail.PublishMessage(es)
			if errOne != nil {
				fmt.Println("error publishing email to queue ", errOne)
			}
			fmt.Println("error connecting to IPFS network ", err)
			d.Ack(false)
			continue
		}
		err = ipfsManager.Shell.Unpin(rm.ContentHash)
		if err != nil {
			addresses := []string{rm.EthAddress}
			es := EmailSend{
				Subject:      "Pin removal failed",
				Content:      fmt.Sprintf("Pin removal failed for ipfs network %s due to reason %s", rm.NetworkName, err),
				ContentType:  "",
				EthAddresses: addresses,
			}
			errOne := qmEmail.PublishMessage(es)
			if errOne != nil {
				//TODO log and handle
				fmt.Println("error publishing email to queue ", errOne)
			}
			fmt.Println("failed to remove content hash ", err)
			d.Ack(false)
			continue
		}
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
		// TODO: add email notification indicating that the file was added, giving the content hash for the particular file
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
		d.Ack(false)
	}
	return nil
}
