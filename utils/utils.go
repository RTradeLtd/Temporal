package utils

import (
	"errors"
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
	IPNSPublishPublic             = 5.00
	IPNSPublishPrivate            = 10.00
	RSAKeyCreationPublic          = 2.00
	RSAKeyCreationPrivate         = 2.50
	EDKeyCreationPublic           = 1.00
	EDKeyCreationPrivate          = 1.50
	DNSLinkGenerationPublic       = 5.00
	DNSLinkGenerationPrivate      = 5.00
)

// this is a testing parameter for now, exact costs will be detailed at a later time
var usdPerGigabytePerMonth = 0.134

// NilTime is used to compare empty time
var NilTime time.Time

// CalculateAPICallCost is used to calculate the cost associated with an API call,
// that isn't related to uploads/pinning
func CalculateAPICallCost(callType string, privateNetwork bool) (float64, error) {
	var cost float64
	switch callType {
	case "ipns":
		if privateNetwork {
			cost = IPNSPublishPrivate
		} else {
			cost = IPNSPublishPublic
		}
	case "pubsub":
		if privateNetwork {
			cost = PubSubPublishPrivate
		} else {
			cost = PubSubPublishPublic
		}
	case "ed-key":
		if privateNetwork {
			cost = EDKeyCreationPrivate
		} else {
			cost = EDKeyCreationPublic
		}
	case "rsa-key":
		if privateNetwork {
			cost = RSAKeyCreationPrivate
		} else {
			cost = RSAKeyCreationPublic
		}
	case "dlink":
		if privateNetwork {
			cost = DNSLinkGenerationPrivate
		} else {
			cost = DNSLinkGenerationPublic
		}
	default:
		return 0, errors.New("call type unsupported")
	}
	return cost, nil
}

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

func BytesToGigaBytes(size int64) float64 {
	gigabytes := float64(datasize.GB.Bytes())
	sizeInGigaBytes := float64(size) / gigabytes
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
