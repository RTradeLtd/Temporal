package models

import (
	"errors"
	"fmt"

	"github.com/jinzhu/gorm"
)

// Record is an entry within a tns zone
type Record struct {
	gorm.Model
	UserName       string      `gorm:"type:varchar(255)"`
	Name           string      `gorm:"type:varchar(255)"`
	RecordKeyName  string      `gorm:"type:varchar(255)"`
	LatestIPFSHash string      `gorm:"type:varchar(255)"`
	ZoneName       string      `gorm:"type:varchar(255)"`
	MetaData       interface{} `gorm:"type:json"` // we need to parse this to a "string json"
}

// RecordManager is used to manipulate records in our db
type RecordManager struct {
	DB *gorm.DB
}

// NewRecordManager is used to generate our record manager
func NewRecordManager(db *gorm.DB) *RecordManager {
	return &RecordManager{DB: db}
}

// FindRecordByNameAndUser is used to search fro a record by name and user
func (rm *RecordManager) FindRecordByNameAndUser(username, name string) (*Record, error) {
	r := Record{}
	if check := rm.DB.Where("user_name = ? AND name = ?", username, name); check.Error != nil {
		return nil, check.Error
	}
	return &r, nil
}

// AddRecord is used to save a record to our database
func (rm *RecordManager) AddRecord(username, recordName, recordKeyName, zoneName string, metadata map[string]interface{}) (*Record, error) {
	if _, err := rm.FindRecordByNameAndUser(username, recordName); err == nil {
		return nil, errors.New("record already exists")
	}
	r := Record{
		UserName:      username,
		Name:          recordName,
		RecordKeyName: recordKeyName,
		ZoneName:      zoneName,
		MetaData:      rm.stringifyMetaData(metadata),
	}
	if check := rm.DB.Create(&r); check.Error != nil {
		return nil, check.Error
	}
	return &r, nil
}

// StringifyMetaData is sued to convert metadata, into a string object json object
func (rm *RecordManager) stringifyMetaData(data map[string]interface{}) string {
	s := "{"
	count := 0
	for k, v := range data {
		s = fmt.Sprintf("\"%s\": \"%s\"", k, v)
		if count == len(data)-1 {
			s = fmt.Sprintf("%s}", s)
		}
	}
	return s
}
