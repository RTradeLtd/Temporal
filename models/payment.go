package models

import (
	"errors"
	"strconv"

	"github.com/jinzhu/gorm"
)

// Payments is our payment model
type Payments struct {
	gorm.Model
	Number         int64   `gorm:"type:integer"`
	DepositAddress string  `gorm:"type:varchar(255)"`
	TxHash         string  `gorm:"type:varchar(255)"`
	USDValue       float64 `gorm:"type:float"` // USDValue is also a "Credit" value, since 1 USD -> 1 Credit
	Blockchain     string  `gorm:"type:varchar(255)"`
	Type           string  `gorm:"type:varchar(255)"` // ETH, RTC, XMR, BTC, LTC
	UserName       string  `gorm:"type:varchar(255)"`
	Confirmed      bool    `gorm:"type:varchar(255)"`
}

// PaymentManager is used to interact with payment information in our database
type PaymentManager struct {
	DB *gorm.DB
}

// NewPaymentManager is used to generate our payment manager helper
func NewPaymentManager(db *gorm.DB) *PaymentManager {
	return &PaymentManager{DB: db}
}

// FindPaymentByNumber is used to find a payment by its payment number
func (pm *PaymentManager) FindPaymentByNumber(username string, number int64) (*Payments, error) {
	p := Payments{}
	if check := pm.DB.Where("user_name = ? AND number = ?", username, number).First(&p); check.Error != nil {
		return nil, check.Error
	}
	return &p, nil
}

// GetLatestPaymentNumber is used to get the latest payment number for a user
func (pm *PaymentManager) GetLatestPaymentNumber(username string) (int64, error) {
	p := Payments{}
	check := pm.DB.Where("user_name = ?", username).Order("number desc").First(&p)
	if check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return 0, check.Error
	}

	if check.Error == gorm.ErrRecordNotFound {
		return 0, nil
	}
	return p.Number + 1, nil
}

// NewPayment is used to create a payment in our database
func (pm *PaymentManager) NewPayment(number int64, depositAddress string, txHash string, usdValue float64, blockchain string, paymentType string, username string) (*Payments, error) {
	p := Payments{}
	// check for a payment with the number
	check := pm.DB.Where("number = ?", number).First(&p)
	if check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return nil, check.Error
	}
	if check.Error == nil {
		return nil, errors.New("payment with number already exists in database")
	}
	// check for a payment with the tx hash
	check = pm.DB.Where("tx_hash = ?", txHash).First(&p)
	if check.Error != nil && check.Error != gorm.ErrRecordNotFound {
		return nil, check.Error
	}
	if check.Error == nil {
		return nil, errors.New("paymnet with tx hash already exists in database")
	}
	p = Payments{
		DepositAddress: depositAddress,
		TxHash:         txHash,
		USDValue:       usdValue,
		Blockchain:     blockchain,
		Type:           paymentType,
		UserName:       username,
		Confirmed:      false,
		Number:         number,
	}

	if check := pm.DB.Create(&p); check.Error != nil {
		return nil, check.Error
	}

	return &p, nil
}

// ConfirmPayment is used to mark a payment as confirmed
func (pm *PaymentManager) ConfirmPayment(txHash string) (*Payments, error) {
	p, err := pm.FindPaymentByTxHash(txHash)
	if err != nil {
		return nil, err
	}
	p.Confirmed = true
	if check := pm.DB.Model(p).Update("confirmed", p.Confirmed); check.Error != nil {
		return nil, check.Error
	}
	return p, nil
}

// FindPaymentByTxHash is used to find a payment by its tx hash
func (pm *PaymentManager) FindPaymentByTxHash(txHash string) (*Payments, error) {
	p := Payments{}
	if check := pm.DB.Where("tx_hash = ?", txHash).First(&p); check.Error != nil {
		return nil, check.Error
	}
	return &p, nil
}

// UpdatePaymentTxHash UpdatePaymentTxHash is used to update the tx hash of a payment
func (pm *PaymentManager) UpdatePaymentTxHash(username, txHash string, number int64) (*Payments, error) {
	payment, err := pm.FindPaymentByNumber(username, number)
	if err != nil {
		return nil, err
	}
	numberString := strconv.FormatInt(number, 10)
	if payment.TxHash != numberString {
		return nil, errors.New("payment already has an updated tx hash")
	}
	payment.TxHash = txHash
	if check := pm.DB.Model(payment).Update("tx_hash", payment.TxHash); check.Error != nil {
		return nil, check.Error
	}
	return payment, nil
}
