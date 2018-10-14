package queue

import (
	"encoding/json"

	"github.com/RTradeLtd/Temporal/rtfs"

	"github.com/RTradeLtd/Temporal/models"
	"github.com/RTradeLtd/Temporal/tns"
	"github.com/RTradeLtd/config"
	"github.com/jinzhu/gorm"
	peer "github.com/libp2p/go-libp2p-peer"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

// ProcessTNSRecordCreation is used to process new TNS record creation requests
func (qm *QueueManager) ProcessTNSRecordCreation(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	zm := models.NewZoneManager(db)
	rm := models.NewRecordManager(db)
	qm.Logger.WithFields(log.Fields{
		"service": qm.Service,
	}).Info("processing messages")
	// process new messages
	for d := range msgs {
		// message received
		qm.Logger.WithFields(log.Fields{
			"service": qm.Service,
		}).Info("new message received")
		req := RecordCreation{}
		// unmarshal message
		if err := json.Unmarshal(d.Body, &req); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}
		// search for zone in db
		if _, err := zm.FindZoneByNameAndUser(req.ZoneName, req.UserName); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("unable to find zone")
			d.Ack(false)
			continue
		}
		// connect to ipfs
		rtfsManager, err := rtfs.Initialize("", "")
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to initialize connection to ipfs")
			d.Ack(false)
			continue
		}
		// create our keystore manager
		if err = rtfsManager.CreateKeystoreManager(); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to initialize keystore manager")
			d.Ack(false)
			continue
		}
		// get private key for record
		recordPK, err := rtfsManager.KeystoreManager.GetPrivateKeyByName(req.RecordKeyName)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to get record private key")
			d.Ack(false)
			continue
		}
		// get id
		recordPKID, err := peer.IDFromPublicKey(recordPK.GetPublic())
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to get id from public key")
			d.Ack(false)
			continue
		}
		// create record object
		r := tns.Record{
			PublicKey: recordPKID.String(),
			Name:      req.RecordName,
			MetaData:  req.MetaData,
		}
		// marshal it
		marshaled, err := json.Marshal(&r)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to marshal tns record")
			d.Ack(false)
			continue
		}
		// put to ipfs
		resp, err := rtfsManager.Shell.DagPut(marshaled, "json", "cbor")
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to put record file to ipfs")
			d.Ack(false)
		}
		// update the zone in database
		zone, err := zm.AddRecordForZone(
			req.ZoneName, req.RecordName, req.UserName,
		)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("unable to add record to zone")
			d.Ack(false)
			continue
		}
		// update the database with a new record
		if _, err := rm.AddRecord(
			req.UserName, req.RecordName, req.RecordKeyName, req.ZoneName, req.MetaData,
		); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("unable to add record to database")
			d.Ack(false)
			continue
		}
		// update the latest ipfs hash for this record
		if _, err := rm.UpdateLatestIPFSHash(
			req.UserName, req.RecordName, resp,
		); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("unable to update ipfs hash for record")
			d.Ack(false)
			continue
		}
		zonePK, err := rtfsManager.KeystoreManager.GetPrivateKeyByName(zone.ZonePublicKeyName)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("unable to get zone pk")
			d.Ack(false)
			continue
		}
		// convert private key to id
		zonePKID, err := peer.IDFromPublicKey(zonePK.GetPublic())
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to get peer id from public key")
			d.Ack(false)
			continue
		}
		// get zone manager private key
		zoneManagerPK, err := rtfsManager.KeystoreManager.GetPrivateKeyByName(zone.ManagerPublicKeyName)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to get zone manager private key")
			d.Ack(false)
			continue
		}
		zomeManagerPKID, err := peer.IDFromPublicKey(zoneManagerPK.GetPublic())
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to get zone manager id from private key")
			d.Ack(false)
			continue
		}
		records, err := rm.FindRecordsByZone(zone.UserName, zone.Name)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to find records")
			d.Ack(false)
			continue
		}
		m := make(map[string]*tns.Record)
		mr := make(map[string]string)
		for _, v := range *records {
			tnR := &tns.Record{
				PublicKey: v.RecordKeyName,
				Name:      v.Name,
				MetaData:  nil,
			}
			m[v.Name] = tnR
			mr[v.Name] = v.RecordKeyName
		}
		z := tns.Zone{
			PublicKey: zonePKID.String(),
			Manager: &tns.ZoneManager{
				PublicKey: zomeManagerPKID.String(),
			},
			Name:                    zone.Name,
			Records:                 m,
			RecordNamesToPublicKeys: mr,
		}
		// marshal to bytes
		marshaled, err = json.Marshal(&z)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to marshaled tns zone")
			d.Ack(false)
			continue
		}
		// put to ipfs
		resp, err = rtfsManager.Shell.DagPut(marshaled, "json", "cbor")
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to put zone file to ipfs")
			d.Ack(false)
			continue
		}
		// update database with has
		zone.LatestIPFSHash = resp
		if _, err = zm.UpdateLatestIPFSHashForZone(zone.Name, zone.UserName, resp); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to update zone in database")
			d.Ack(false)
			continue
		}
		qm.Logger.WithFields(log.Fields{
			"service": qm.Service,
		}).Info("record added to zone")
		d.Ack(false)
	}
	return nil
}

// ProcessTNSZoneCreation is used to process new TNS zone creation requests
func (qm *QueueManager) ProcessTNSZoneCreation(msgs <-chan amqp.Delivery, db *gorm.DB, cfg *config.TemporalConfig) error {
	zm := models.NewZoneManager(db)
	qm.Logger.WithFields(log.Fields{
		"service": qm.Service,
	}).Info("processing messages")
	// process messages
	for d := range msgs {
		// new message
		qm.Logger.WithFields(log.Fields{
			"service": qm.Service,
		}).Info("new message received")
		req := ZoneCreation{}
		// unmarshal the message into a typed format
		if err := json.Unmarshal(d.Body, &req); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to unmarshal message")
			d.Ack(false)
			continue
		}
		// get the zone from db
		zone, err := zm.FindZoneByNameAndUser(req.Name, req.UserName)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to search for zone")
			d.Ack(false)
			continue
		}
		// connect to ipfs
		rtfsManager, err := rtfs.Initialize("", "")
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to intiialize connection to ipfs")
			d.Ack(false)
			continue
		}
		// create keystore manager
		if err = rtfsManager.CreateKeystoreManager(); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to initialize keystore manager")
			d.Ack(false)
			continue
		}
		// get zone manager private key
		zoneManagerPK, err := rtfsManager.KeystoreManager.GetPrivateKeyByName(req.ManagerKeyName)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to get zone manager private key")
			d.Ack(false)
			continue
		}
		// get zone private key
		zonePK, err := rtfsManager.KeystoreManager.GetPrivateKeyByName(req.ZoneKeyName)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to initialize keystore manager")
			d.Ack(false)
			continue
		}
		// convert private key to id
		zonePKID, err := peer.IDFromPublicKey(zonePK.GetPublic())
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to get peer id from public key")
			d.Ack(false)
			continue
		}
		zoneManagerPKID, err := peer.IDFromPublicKey(zoneManagerPK.GetPublic())
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to get peer id from public key")
			d.Ack(false)
			continue
		}
		// generate initial zone object
		z := tns.Zone{
			PublicKey: zonePKID.String(),
			Manager: &tns.ZoneManager{
				PublicKey: zoneManagerPKID.String(),
			},
			Name: req.Name,
		}
		// marshal to bytes
		marshaled, err := json.Marshal(&z)
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to marshaled tns zone")
			d.Ack(false)
			continue
		}
		// put to ipfs
		resp, err := rtfsManager.Shell.DagPut(marshaled, "json", "cbor")
		if err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to put zone file to ipfs")
			d.Ack(false)
			continue
		}
		// update database with has
		zone.LatestIPFSHash = resp
		if _, err = zm.UpdateLatestIPFSHashForZone(zone.Name, zone.UserName, resp); err != nil {
			qm.Logger.WithFields(log.Fields{
				"service": qm.Service,
				"error":   err.Error(),
			}).Error("failed to update zone in database")
			d.Ack(false)
			continue
		}
		// success
		qm.Logger.WithFields(log.Fields{
			"service": qm.Service,
		}).Info("zone published and database is updated")
		d.Ack(false)
		continue
	}
	return nil
}
