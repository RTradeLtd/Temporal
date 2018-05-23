package models

import "github.com/jinzhu/gorm"

type Payment struct {
	gorm.Model
	UploaderAddress string `gorm:"type:varchar(255);"`
	CID             string `gorm:"type:varchar(255);"`
	HashedCID       string `gorm:"type:varchar(255);"`
	PaymentID       string `gorm:"type:varchar(255);"`
	Paid            bool   `gorm:"type:boolean;"`
}

type PaymentManager struct {
	DB *gorm.DB
}

func NewPaymentManager(db *gorm.DB) *PaymentManager {
	return &PaymentManager{DB: db}
}

func (pm *PaymentManager) FindPaymentsByUploader(address string) *[]Payment {
	var payments []Payment
	pm.DB.Find(&payments).Where("uploader_address = ?", address)
	return &payments
}

func (pm *PaymentManager) FindPaymentByPaymentID(paymentID string) *Payment {
	var payment Payment
	pm.DB.Find(&payment).Where("payment_id = ?", paymentID)
	return &payment
}
