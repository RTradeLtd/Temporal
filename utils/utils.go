package utils

import (
	"math/big"
	"time"

	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
	"github.com/c2h5oh/datasize"
)

// prices listed here are temporary and will be changed
const (
	UsdPerGigaBytePerMonthPublic  = 0.134
	UsdPerGigaBytePerMonthPrivate = 0.154
	PubSubPublishPublic           = 0.01
	PubSubPublishPrivate          = 0.02
	IPNSPublishPrivate            = 10.00
	IPNSPublishPublic             = 5.00
)

// this is a testing parameter for now, exact costs will be detailed at a later time
var usdPerGigabytePerMonth = 0.134

// NilTime is used to compare empty time
var NilTime time.Time

// CalculatePinCost is used to calculate the cost of pining a particular content hash
func CalculatePinCost(contentHash string, holdTimeInMonths int64, shell *ipfsapi.Shell, privateNetwork bool) (float64, error) {
	objectStat, err := shell.ObjectStat(contentHash)
	if err != nil {
		return float64(0), err
	}
	// get total size of content hash in bytes
	sizeInBytes := objectStat.CumulativeSize
	// get gigabytes convert to bytes
	gigaInBytes := datasize.GB.Bytes()
	// convert size of content hash form int to float64
	sizeInBytesFloat := float64(sizeInBytes)
	// convert gigabytes to float
	gigabytesFloat := float64(gigaInBytes)
	// convert object size from bytes to gigabytes
	objectSizeInGigabytesFloat := sizeInBytesFloat / gigabytesFloat
	var costPerMonthFloat float64
	if privateNetwork {
		costPerMonthFloat = objectSizeInGigabytesFloat * UsdPerGigaBytePerMonthPrivate
	} else {
		costPerMonthFloat = objectSizeInGigabytesFloat * UsdPerGigaBytePerMonthPublic
	}
	totalCostFloat := costPerMonthFloat * float64(holdTimeInMonths)
	return totalCostFloat, nil
}

func CalculateFileCost(holdTimeInMonths, size int64, privateNetwork bool) float64 {
	gigabytesFloat := float64(datasize.GB.Bytes())
	sizeFloat := float64(size)
	sizeGigabytesFloat := sizeFloat / gigabytesFloat
	var costPerMonthFloat float64
	if privateNetwork {
		costPerMonthFloat = sizeGigabytesFloat * UsdPerGigaBytePerMonthPrivate
	} else {
		costPerMonthFloat = sizeGigabytesFloat * UsdPerGigaBytePerMonthPublic
	}
	totalCostUSDFloat := costPerMonthFloat * float64(holdTimeInMonths)
	return totalCostUSDFloat
}

func BytesToGigaBytes(size int64) int64 {
	gigabytes := int64(datasize.GB.Bytes())
	sizeInGigaBytes := size / gigabytes
	return sizeInGigaBytes
}

// FloatToBigInt used to convert a float to big int
func FloatToBigInt(val float64) *big.Int {
	bigval := new(big.Float)
	bigval.SetFloat64(val)
	// Set precision if required.
	// bigval.SetPrec(64)

	coin := new(big.Float)
	coin.SetInt(big.NewInt(1000000000000000000))

	bigval.Mul(bigval, coin)

	result := new(big.Int)
	bigval.Int(result) // store converted number in result

	return result
}
