package models

import (
	"errors"

	"github.com/c2h5oh/datasize"

	"github.com/RTradeLtd/gorm"
)

// DataUsageTier is a type of usage tier
// which governs the price per gb ratio
type DataUsageTier string

// String returns the value of DataUsageTier as a string
func (d DataUsageTier) String() string {
	return string(d)
}

// PricePerGB returns the price per gb of a usage tier
func (d DataUsageTier) PricePerGB() float64 {
	switch d {
	case Light:
		return 0.22
	case Plus:
		return 0.165
	case Partner:
		return 0.16
	default:
		// this is a catch-all for free tier
		// free tier users will never encounter a charge call
		return 9999
	}
}

var (
	// Free is what every signed up user is automatically registered as
	// Restrictions of free:
	//			* No on-demand data encryption
	//			* 3GB/month max
	//			* IPNS limit of 5, with no automatic republishes
	//			* 5 keys
	Free DataUsageTier = "free"

	// Partner is for partners of RTrade
	// partners have 100GB/month free
	//			* on-demand data encryption
	//			* 0.16GB/month after 100GB limit
	Partner DataUsageTier = "partner"

	// Light is the first non-free, and non-partner tier
	// tier is from 3GB -> 100GB
	// after reaching 100GB they are upgraded to Plus
	//			* on-demand data encryption
	//			* 0.22
	Light DataUsageTier = "light"

	// Plus is the other non-free, non-partner, non-light tier
	// tier is from 100GB -> 1TB to our max monthly usage of 1TB
	// 			* on-demand data encryption
	//			* $0.165
	Plus DataUsageTier = "plus"

	// FreeUploadLimit is the maximum data usage for free accounts
	// Currrently set to 3GB
	FreeUploadLimit = 3 * datasize.GB.Bytes()

	// NonFreeUploadLimit is the maximum data usage for non-free accounts
	// Currently set to 1TB
	NonFreeUploadLimit = datasize.TB.Bytes()

	// PlusTierMinimumUpload is the current data usage
	// needed to upgrade from Light -> Plus
	// currently set to 100GB
	PlusTierMinimumUpload = 100 * datasize.GB.Bytes()

	// FreeKeyLimit defines how many keys free accounts can create
	FreeKeyLimit int64 = 5
	// FreePubSubLimit defines how many pubsub messages free accounts can send
	FreePubSubLimit int64 = 100
	// FreeIPNSLimit defines how many ipns records free accounts can publish
	FreeIPNSLimit int64 = 5

	// PartnerKeyLimit defines how many keys partner accounts can create
	PartnerKeyLimit int64 = 200
	// PartnerPubSubLimit defines how many pubsub messages partner accounts can send
	PartnerPubSubLimit int64 = 20000
	// PartnerIPNSLimit defines how many ipns records partner accounts can publish
	PartnerIPNSLimit int64 = 200

	// LightKeyLimit defines how many keys light accounts can create
	LightKeyLimit int64 = 100
	// LightPubSubLimit defines how many pubsub messages light accounts can send
	LightPubSubLimit int64 = 10000
	// LightIPNSLimit defines how many ipns records light accounts can publish
	LightIPNSLimit int64 = 100

	// PlusKeyLimit defines how many keys plus accounts can create
	PlusKeyLimit int64 = 150
	// PlusPubSubLimit defines how many pubsub messages plus accounts can send
	PlusPubSubLimit int64 = 15000
	// PlusIPNSRecordLimit defines how many ipns records plus accounts can publish
	PlusIPNSRecordLimit int64 = 150
)

// Usage is used to handle Usage of Temporal accounts
type Usage struct {
	gorm.Model
	UserName string `gorm:"type:varchar(255);unique"`
	// keeps track of the max monthly upload limit for the user
	MonthlyDataLimitBytes uint64 `gorm:"type:numeric;default:0"`
	// keeps track of the current monthyl upload limit used
	CurrentDataUsedBytes uint64 `gorm:"type:numeric;default:0"`
	// keeps track of how many IPNS records the user has published
	IPNSRecordsPublished int64 `gorm:"type:integer;default:0"`
	// keeps track of how many ipns records the user is allowed to publish
	IPNSRecordsAllowed int64 `gorm:"type:integer;default:0"`
	// keeps track of how many messages the user has sent
	PubSubMessagesSent int64 `gorm:"type:integer;default:0"`
	// keeps track of the number of pubsub messages a user is allowed to send
	PubSubMessagesAllowed int64 `gorm:"type:integer;default:0"`
	// keeps track of how many keys the user has created
	KeysCreated int64 `gorm:"type:integer;default:0"`
	// keeps track of how many keys the user is allowed to create
	KeysAllowed int64 `gorm:"type:integer;default:0"`
	// keeps track of the tier the user belongs to
	Tier DataUsageTier `gorm:"type:varchar(255)"`
}

// UsageManager is used to manage Usage models
type UsageManager struct {
	DB *gorm.DB
}

// NewUsageManager is used to instantiate a Usage manager
func NewUsageManager(db *gorm.DB) *UsageManager {
	return &UsageManager{DB: db}
}

// NewUsageEntry is used to create a new usage entry in our database
// if tier is free, limit to 3GB monthly otherwise set to 1TB
func (bm *UsageManager) NewUsageEntry(username string, tier DataUsageTier) (*Usage, error) {
	usage := &Usage{
		UserName:             username,
		CurrentDataUsedBytes: 0,
		IPNSRecordsPublished: 0,
		PubSubMessagesSent:   0,
		Tier:                 tier,
	}
	// set tier
	usage.Tier = tier
	// set tier based restrictions
	switch tier {
	case Free:
		usage.MonthlyDataLimitBytes = FreeUploadLimit
		usage.KeysAllowed = FreeKeyLimit
		usage.PubSubMessagesAllowed = FreePubSubLimit
		usage.IPNSRecordsAllowed = FreeIPNSLimit
	case Partner:
		usage.MonthlyDataLimitBytes = NonFreeUploadLimit
		usage.KeysAllowed = PartnerKeyLimit
		usage.PubSubMessagesAllowed = PartnerPubSubLimit
		usage.IPNSRecordsAllowed = PartnerIPNSLimit
	case Light:
		usage.MonthlyDataLimitBytes = NonFreeUploadLimit
		usage.KeysAllowed = LightKeyLimit
		usage.PubSubMessagesAllowed = LightPubSubLimit
		usage.IPNSRecordsAllowed = LightIPNSLimit
	case Plus:
		usage.MonthlyDataLimitBytes = NonFreeUploadLimit
		usage.KeysAllowed = PlusKeyLimit
		usage.PubSubMessagesAllowed = PlusPubSubLimit
		usage.IPNSRecordsAllowed = PlusIPNSRecordLimit
	default:
		return nil, errors.New("unsupported tier provided")
	}
	if err := bm.DB.Create(usage).Error; err != nil {
		return nil, err
	}
	return usage, nil
}

// FindByUserName is used to find a Usage model by the associated username
func (bm *UsageManager) FindByUserName(username string) (*Usage, error) {
	b := Usage{}
	if check := bm.DB.Where("user_name = ?", username).First(&b); check.Error != nil {
		return nil, check.Error
	}
	return &b, nil
}

// GetUploadPricePerGB is used to get the upload price per gb for a user
// allows us to specify whether the payment
func (bm *UsageManager) GetUploadPricePerGB(username string) (float64, error) {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return 0, err
	}
	return b.Tier.PricePerGB(), nil
}

// CanPublishIPNS is used to check if a user can publish IPNS records
func (bm *UsageManager) CanPublishIPNS(username string) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	if b.IPNSRecordsPublished >= b.IPNSRecordsAllowed {
		return errors.New("too many records published")
	}
	return nil
}

// CanUpload is used to check if a user can upload an object with the given data size
func (bm *UsageManager) CanUpload(username string, dataSizeBytes uint64) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	if b.CurrentDataUsedBytes+dataSizeBytes > b.MonthlyDataLimitBytes {
		return errors.New("upload will breach max monthly data usage, please upload a smaller file")
	}
	return nil
}

// CanPublishPubSub is used to check if a user can publish pubsub messages
func (bm *UsageManager) CanPublishPubSub(username string) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	if b.PubSubMessagesSent >= b.PubSubMessagesAllowed {
		return errors.New("too many pubsub messages sent, please wait until next billing cycle")
	}
	return nil
}

// CanCreateKey is used to check if a user can create an ipfs key
func (bm *UsageManager) CanCreateKey(username string) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	if b.KeysCreated >= b.KeysAllowed {
		return errors.New("too many keys created, please wait until next billing cycle")
	}
	return nil
}

// UpdateDataUsage is used to update the users' data usage amount
// If the account is non free, and the upload pushes their total monthly usage
// above the tier limit, they will be upgraded to the next tier to receive the discounted price
// the discounted price will apply on subsequent uploads.
// If the 1TB maximum monthly limit is hit, then we throw an error
func (bm *UsageManager) UpdateDataUsage(username string, uploadSizeBytes uint64) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	// update total data used
	b.CurrentDataUsedBytes = b.CurrentDataUsedBytes + uploadSizeBytes
	// perform a tier check for light accounts
	// if they use more than 100GB, upgrade them to Plus tier
	if b.Tier == Light {
		// if they are light plan, and this upload takes them over 100GB
		// update their tier to plus, enabling cheaper data rates
		if b.CurrentDataUsedBytes >= PlusTierMinimumUpload {
			// update tier
			b.Tier = Plus
			b.KeysAllowed = PlusKeyLimit
			b.PubSubMessagesAllowed = PlusPubSubLimit
			b.IPNSRecordsAllowed = PlusIPNSRecordLimit
		}
	}
	// perform upload limit checks
	if b.Tier == Free {
		// if they are free, they will need to upgrade their plan
		if b.CurrentDataUsedBytes >= FreeUploadLimit {
			return errors.New("upload limit will be reached, please upload smaller content or upgrade your plan")
		}
	} else {
		// check for the max upload limit of 1TB
		if b.CurrentDataUsedBytes >= NonFreeUploadLimit {
			return errors.New("max upload limit of 1TB reached, contact support")
		}
	}
	// save updated columns and return
	return bm.DB.Model(b).UpdateColumns(map[string]interface{}{
		"tier":                     b.Tier,
		"current_data_used_bytes":  b.CurrentDataUsedBytes,
		"keys_allowed":             b.KeysAllowed,
		"ip_ns_records_allowed":    b.IPNSRecordsAllowed,
		"pub_sub_messages_allowed": b.PubSubMessagesAllowed,
	}).Error
}

// ReduceDataUsage is used to reduce a users current data used. This is used in cases
// where processing within the queue system fails, and we need to reset their data usage
func (bm *UsageManager) ReduceDataUsage(username string, uploadSizeBytes uint64) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	// reduce total data used
	// if the current data used is smaller than the reduction size
	// reset their data used to 0
	if b.CurrentDataUsedBytes < uploadSizeBytes {
		b.CurrentDataUsedBytes = 0
	} else {
		b.CurrentDataUsedBytes = b.CurrentDataUsedBytes - uploadSizeBytes
	}
	// perform tier downgrade check
	// accounts can never be downgraded below Light to free tier
	if b.Tier == Plus {
		if b.CurrentDataUsedBytes < PlusTierMinimumUpload {
			b.Tier = Light
			b.KeysAllowed = LightKeyLimit
			b.IPNSRecordsAllowed = LightIPNSLimit
			b.PubSubMessagesAllowed = LightPubSubLimit
		}
	}
	return bm.DB.Model(b).UpdateColumns(map[string]interface{}{
		"tier":                    b.Tier,
		"current_data_used_Bytes": b.CurrentDataUsedBytes,
	}).Error
}

// ReduceKeyCount is used to reduce the number of keys a user has created
func (bm *UsageManager) ReduceKeyCount(username string, count int64) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	if b.KeysCreated > count {
		b.KeysCreated = 0
	} else {
		b.KeysCreated = b.KeysCreated - count
	}
	return bm.DB.Model(b).Update("keys_created", b.KeysCreated).Error
}

// UpdateTier is used to update the Usage tier associated with an account
// accounts may never be downgraded back to Free
func (bm *UsageManager) UpdateTier(username string, tier DataUsageTier) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	// set tier
	b.Tier = tier
	// set tier based restrictions
	switch tier {
	case Partner:
		b.MonthlyDataLimitBytes = NonFreeUploadLimit
		b.KeysAllowed = PartnerKeyLimit
		b.PubSubMessagesAllowed = PartnerPubSubLimit
		b.IPNSRecordsAllowed = PartnerIPNSLimit
	case Light:
		b.MonthlyDataLimitBytes = NonFreeUploadLimit
		b.KeysAllowed = LightKeyLimit
		b.PubSubMessagesAllowed = LightPubSubLimit
		b.IPNSRecordsAllowed = LightIPNSLimit
	case Plus:
		b.MonthlyDataLimitBytes = NonFreeUploadLimit
		b.KeysAllowed = PlusKeyLimit
		b.PubSubMessagesAllowed = PlusPubSubLimit
		b.IPNSRecordsAllowed = PlusIPNSRecordLimit
	default:
		return errors.New("unsupported tier provided")
	}

	return bm.DB.Model(b).UpdateColumns(map[string]interface{}{
		"tier":                     b.Tier,
		"keys_allowed":             b.KeysAllowed,
		"pub_sub_messages_allowed": b.PubSubMessagesAllowed,
		"ip_ns_records_allowed":    b.IPNSRecordsAllowed,
		"monthly_data_limit_bytes": b.MonthlyDataLimitBytes},
	).Error
}

// IncrementPubSubUsage is used to increment the pubsub publish counter
func (bm *UsageManager) IncrementPubSubUsage(username string, count int64) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	b.PubSubMessagesSent = b.PubSubMessagesSent + count
	return bm.DB.Model(b).Update("pub_sub_messages_sent", b.PubSubMessagesSent).Error
}

// IncrementIPNSUsage is used to increment the ipns record publish counter
func (bm *UsageManager) IncrementIPNSUsage(username string, count int64) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	b.IPNSRecordsPublished = b.IPNSRecordsPublished + count
	return bm.DB.Model(b).Update("ip_ns_records_published", b.IPNSRecordsPublished).Error
}

// IncrementKeyCount is used to increment the key created counter
func (bm *UsageManager) IncrementKeyCount(username string, count int64) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	b.KeysCreated = b.KeysCreated + count
	return bm.DB.Model(b).Update("keys_created", b.KeysCreated).Error
}

// ResetCounts is used to reset monthly usage counts.
// This does not apply to keys, as keys are a fixed limitation
func (bm *UsageManager) ResetCounts(username string) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	b.CurrentDataUsedBytes = 0
	b.IPNSRecordsPublished = 0
	b.PubSubMessagesSent = 0
	return bm.DB.Model(b).UpdateColumns(map[string]interface{}{
		"current_data_used_bytes": b.CurrentDataUsedBytes,
		"ip_ns_records_published": b.IPNSRecordsPublished,
		"pub_sub_messages_sent":   b.PubSubMessagesSent,
	}).Error
}
