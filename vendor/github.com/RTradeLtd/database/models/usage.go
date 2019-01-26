package models

import (
	"errors"
	"time"

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
	FreeUploadLimit = datasize.GB.Bytes() * 3

	// NonFreeUploadLimit is the maximum data usage for non-free accounts
	// Currently set to 1TB
	NonFreeUploadLimit = datasize.TB.Bytes()

	// PlusTierMinimumUpload is the current data usage
	// needed to upgrade from Light -> Plus
	// currently set to 100GB
	PlusTierMinimumUpload = datasize.GB.Bytes() * 100
)

// Usage is used to handle Usage of Temporal accounts
type Usage struct {
	UserName string `gorm:"type:varchar(255);unique"`
	// keeps track of the max monthly upload limit for the user
	MonthlyDataLimitGB uint64 `gorm:"type:numeric;default:0"`
	// keeps track of the current monthyl upload limit used
	CurrentDataUsedGB uint64 `gorm:"type:numeric;default:0"`
	// keeps track of how many IPNS records the user has published
	IPNSRecordsPublished int64 `gorm:"type:integer;default:0"`
	// keeps track of how many messages the user has sent
	PubSubMessagesSent int64 `gorm:"type:integer;default:0"`
	// Used to indicate whether or not the user
	// has consumed their free private network trial
	PrivateNetworkTrialUsed bool `gorm:"type:boolean;default:false"`
	// Used to determine when their private network trial ends
	// good until 2038 <- ticket is open in JIRA so we do not forget about this
	// trial is 800 hours in unix timestamp
	TrialEndTime int64 `gorm:"type:integer;default:0"`
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
		UserName:                username,
		CurrentDataUsedGB:       0,
		IPNSRecordsPublished:    0,
		PubSubMessagesSent:      0,
		PrivateNetworkTrialUsed: false,
		TrialEndTime:            0,
		Tier:                    tier,
	}
	if tier == Free {
		usage.MonthlyDataLimitGB = FreeUploadLimit
	} else {
		usage.MonthlyDataLimitGB = NonFreeUploadLimit
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
func (bm *UsageManager) CanPublishIPNS(username string) (bool, error) {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return false, err
	}
	// if they are free, make sure they aren't at the limit
	if b.Tier == Free && b.IPNSRecordsPublished > 5 {
		return false, nil
	}
	return true, nil
}

// CanUpload is used to check if a user can upload an object with the given data size
func (bm *UsageManager) CanUpload(username string, dataSizeBytes uint64) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	if b.CurrentDataUsedGB+dataSizeBytes > b.MonthlyDataLimitGB {
		return errors.New("upload will breach max monthly data usage, please upload a smaller file")
	}
	return nil
}

// CanPublishPubSub is used to check if a user can publish pubsub messages
func (bm *UsageManager) CanPublishPubSub(username string) (bool, error) {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return false, err
	}
	if b.Tier == Free && b.IPNSRecordsPublished > 10000 {
		return false, nil
	}
	return true, nil
}

// HasStartedPrivateNetworkTrial is used to check if the user has started their private network trial
func (bm *UsageManager) HasStartedPrivateNetworkTrial(username string) (bool, error) {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return false, err
	}
	return b.PrivateNetworkTrialUsed, nil
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
	b.CurrentDataUsedGB = b.CurrentDataUsedGB + uploadSizeBytes
	// perform a tier check for light accounts
	// if they use more than 100GB, upgrade them to Plus tier
	if b.Tier == Light {
		// if they are light plan, and this upload takes them over 100GB
		// update their tier to plus, enabling cheaper data rates
		if b.CurrentDataUsedGB >= PlusTierMinimumUpload {
			// update tier
			b.Tier = Plus
		}
	}
	// perform upload limit checks
	if b.Tier == Free {
		// if they are free, they will need to upgrade their plan
		if b.CurrentDataUsedGB >= FreeUploadLimit {
			return errors.New("upload limit will be reached, please upload smaller content or upgrade your plan")
		}
	} else {
		// check for the max upload limit of 1TB
		if b.CurrentDataUsedGB >= NonFreeUploadLimit {
			return errors.New("max upload limit of 1TB reached, contact support")
		}
	}
	// save updated columns and return
	return bm.DB.Model(b).UpdateColumns(map[string]interface{}{
		"tier":                 b.Tier,
		"current_data_used_gb": b.CurrentDataUsedGB,
	}).Error
}

// UpdateTier is used to update the Usage tier associated with an account
func (bm *UsageManager) UpdateTier(username string, tier DataUsageTier) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	switch tier {
	case Partner:
		b.Tier = Partner
	case Light:
		b.Tier = Light
	case Plus:
		b.Tier = Plus
	}
	return bm.DB.Model(b).Update("tier", b.Tier).Error
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

// StartPrivateNetworkTrial is used to start a users private network trial
func (bm *UsageManager) StartPrivateNetworkTrial(username string) error {
	b, err := bm.FindByUserName(username)
	if err != nil {
		return err
	}
	alreadyStarted, err := bm.HasStartedPrivateNetworkTrial(username)
	if err != nil {
		return err
	}
	if alreadyStarted {
		return errors.New("user has already started their private network trial")
	}
	// update trial end time
	b.TrialEndTime = time.Now().Add(time.Hour * 800).Unix()
	// mark trial as started
	b.PrivateNetworkTrialUsed = true
	// update user model and return error
	return bm.DB.Model(b).UpdateColumns(map[string]interface{}{
		"private_network_trial_used": b.PrivateNetworkTrialUsed,
		"trial_end_time":             b.TrialEndTime}).Error
}
