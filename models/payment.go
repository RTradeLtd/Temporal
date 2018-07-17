package models

import (
	"math/big"

	"github.com/jinzhu/gorm"
)

type PinPayment struct {
	gorm.Model
	Method          uint8
	Number          *big.Int
	ChargeAmount    *big.Int
	UploaderAddress string
	ContentHash     string
}

type PinPaymentManager struct {
	DB *gorm.DB
}

func NewPinPaymentManager(db *gorm.DB) *PinPaymentManager {
	return &PinPaymentManager{DB: db}
}

func (ppm *PinPaymentManager) NewPayment(method uint8, number, chargeAmount *big.Int, uploaderAddress, contentHash string) (*PinPayment, error) {
	pp := &PinPayment{
		Number:          number,
		Method:          method,
		ChargeAmount:    chargeAmount,
		UploaderAddress: uploaderAddress,
		ContentHash:     contentHash,
	}
	if check := ppm.DB.Create(pp); check.Error != nil {
		return nil, check.Error
	}
	return pp, nil
}

func (ppm *PinPaymentManager) RetrieveLatestPaymentNumber() (*big.Int, error) {
	pp := PinPayment{}
	check := ppm.DB.Order("number desc").First(&pp)
	if check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return nil, check.Error
	}
	if check.Error == gorm.ErrRecordNotFound {
		return big.NewInt(0), nil
	}
	return pp.Number, nil
}
