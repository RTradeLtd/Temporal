package models

import (
	"errors"
	"math/big"

	"github.com/jinzhu/gorm"
)

type PinPayment struct {
	gorm.Model
	Method           uint8  `json:"method"`
	Number           string `json:"number"`
	ChargeAmount     string `json:"charge_amount"`
	EthAddress       string `json:"eth_address"`
	UserName         string `json:"user_name"`
	ContentHash      string `json:"content_hash"`
	NetworkName      string `json:"network_name"`
	HoldTimeInMonths int64  `json:"hold_time_in_months"`
}

// Payment is our database model for payments
// Note that if it is of type file, then object name is the minio object
// if it is of type pin, then the object name is the content hash
type Payment struct {
	gorm.Model
	Method           uint8  `json:"method"`
	Number           string `json:"number"`
	ChargeAmount     string `json:"charge_amount"`
	EthAddress       string `json:"eth_address"`
	UserName         string `json:"user_name"`
	NetworkName      string `json:"network_name"`
	ObjectName       string `json:"content_hash"`
	Type             string `json:"time"`
	HoldTimeInMonths int64  `json:"hold_time_in_months"`
}
type PaymentManager struct {
	DB *gorm.DB
}

func NewPaymentManager(db *gorm.DB) *PaymentManager {
	return &PaymentManager{DB: db}
}

func (pm *PaymentManager) NewPayment(method uint8, number *big.Int, chargeAmount *big.Int, ethAddress, objectName, username, uploadType, networkName string, holdTimeInMonths int64) (*Payment, error) {
	p := Payment{}
	check := pm.DB.Where("eth_address = ? AND payment_number = ?", ethAddress, number.String()).First(&p)
	if check.Error == nil {
		return nil, errors.New("payment number already in database for address")
	}
	if check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return nil, check.Error
	}
	p.Method = method
	p.Number = number.String()
	p.ChargeAmount = chargeAmount.String()
	p.EthAddress = ethAddress
	p.UserName = username
	p.NetworkName = networkName
	p.ObjectName = objectName
	p.Type = uploadType
	p.HoldTimeInMonths = holdTimeInMonths
	if check = pm.DB.Create(&p); check.Error != nil {
		return nil, check.Error
	}
	return &p, nil
}

func (pm *PaymentManager) FindPaymentByNumberAndAddress(paymentNumber, ethAddress string) (*Payment, error) {
	p := Payment{}
	if check := pm.DB.Where("eth_address = ? AND payment_number = ?", ethAddress, paymentNumber).First(&p); check.Error != nil {
		return nil, check.Error
	}
	return &p, nil
}

func (pm *PaymentManager) RetrieveLatestPaymentForUser(username string) (*Payment, error) {
	p := Payment{}
	if check := pm.DB.Where("user_name = ?", username).Last(&p); check.Error != nil {
		return nil, check.Error
	}
	return &p, nil
}

func (pm *PaymentManager) RetrieveLatestPaymentNumberForUser(username string) (*big.Int, error) {
	p := Payment{}
	if check := pm.DB.Where("user_name = ?", username).Last(&p); check.Error != nil {
		return nil, check.Error
	}
	num, ok := new(big.Int).SetString(p.Number, 10)
	if !ok {
		return nil, errors.New("failed to convert string to big int")
	}
	return num, nil
}
