package utils

import (
	"errors"
	"math/big"
	"time"

	"github.com/RTradeLtd/rtfs"
	"github.com/c2h5oh/datasize"
)

// prices listed here are temporary and will be changed
const (
	// UsdPerGigaBytePerMonthPublic is the usd cost of storing a gigabyte of data in public networks for a month
	UsdPerGigaBytePerMonthPublic = 0.134
	// UsdPerGigaBytePerMonthPrivate is the usd cost of storing a gigabyte of data in private networks for a month
	UsdPerGigaBytePerMonthPrivate = 0.154
	// PubSubPublishPublic is the of cost of sending a pubsub message to the public ipfs network
	PubSubPublishPublic = 0.01
	// PubSubPublishPrivate is the cost of sending a pubsub message to a private ipfs network
	PubSubPublishPrivate = 0.02
	// IPNSPublishPublic is the cost of publishing an IPNS record to the public ipfs network
	IPNSPublishPublic = 5.00
	// IPNSPublishPrivate is the cost of publishing an IPNS record to the a private ipfs network
	IPNSPublishPrivate = 10.00
	// RSAKeyCreationPublic is the cost of creating an rsa key on the public ipfs network
	RSAKeyCreationPublic = 2.00
	// RSAKeyCreationPrivate is the cost of creating an rsa key on a private ifps network
	RSAKeyCreationPrivate = 2.50
	// EDKeyCreationPublic is the cost of creating an ed25519 key on the public ipfs network
	EDKeyCreationPublic = 1.00
	// EDKeyCreationPrivate is the cost of creating an ed25519 key on a private ipfs network
	EDKeyCreationPrivate = 1.50
	// DNSLinkGenerationPublic is the cost of publishing a dnslink entry for the public ipfs network
	DNSLinkGenerationPublic = 5.00
	// DNSLinkGenerationPrivate is the cost of publishign a dnslink entry for a private ipfs network
	DNSLinkGenerationPrivate = 5.00
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
func CalculatePinCost(contentHash string, holdTimeInMonths int64, im *rtfs.Manager, privateNetwork bool) (float64, error) {
	objectStat, err := im.Stat(contentHash)
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

// CalculateFileCost is used to calculate the cost of storing a file
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

// BytesToGigaBytes is used to convert the given bytes to its gigabyte size
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
