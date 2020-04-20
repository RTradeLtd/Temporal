package utils

import (
	"math/big"

	"github.com/RTradeLtd/database/v2/models"
	"github.com/RTradeLtd/rtfs/v2"
	"github.com/c2h5oh/datasize"
)

// CalculatePinCost is used to calculate the cost of pining a particular content hash
// it returns the cost to bill the user, as well as the calculated size of the pin
func CalculatePinCost(username, contentHash string, holdTimeInMonths int64, im rtfs.Manager, um *models.UsageManager) (float64, int64, error) {
	// get total size of content hash in bytes ensuring that we calculate size
	// by following unique references
	sizeInBytes, _, err := rtfs.DedupAndCalculatePinSize(contentHash, im)
	if err != nil {
		return 0, 0, err
	}
	// if this is true, fall back to default calculation
	// as it wont always be possible to calculate deduplicated
	// storage costs if the object is not of a unixfs type
	if sizeInBytes <= 0 {
		stats, err := im.Stat(contentHash)
		if err != nil {
			return 0, 0, err
		}
		sizeInBytes = int64(stats.CumulativeSize)
	}
	// get gigabytes convert to bytes
	gigaInBytes := datasize.GB.Bytes()
	// convert size of content hash form int to float64
	sizeInBytesFloat := float64(sizeInBytes)
	// convert gigabytes to float
	gigabytesFloat := float64(gigaInBytes)
	// convert object size from bytes to gigabytes
	objectSizeInGigabytesFloat := sizeInBytesFloat / gigabytesFloat
	// get the users usage model
	usage, err := um.FindByUserName(username)
	if err != nil {
		return 0, 0, err
	}
	// if they are free tier, they don't incur data charges
	if usage.Tier == models.Free || usage.Tier == models.WhiteLabeled || usage.Tier == models.Unverified {
		return 0, sizeInBytes, nil
	}
	// dynamic pricing based on their usage tier
	costPerMonthFloat := objectSizeInGigabytesFloat * usage.Tier.PricePerGB()
	return costPerMonthFloat * float64(holdTimeInMonths), sizeInBytes, nil
}

// CalculateFileCost is used to calculate the cost of storing a file
func CalculateFileCost(username string, holdTimeInMonths, size int64, um *models.UsageManager) (float64, error) {
	gigabytesFloat := float64(datasize.GB.Bytes())
	sizeFloat := float64(size)
	sizeGigabytesFloat := sizeFloat / gigabytesFloat
	// get the users usage model
	usage, err := um.FindByUserName(username)
	if err != nil {
		return 0, err
	}
	// if they are free tier, they don't incur data charges
	if usage.Tier == models.Free || usage.Tier == models.WhiteLabeled || usage.Tier == models.Unverified {
		return 0, nil
	}
	// dynamic pricing based on their usage tier
	costPerMonthFloat := sizeGigabytesFloat * usage.Tier.PricePerGB()
	return costPerMonthFloat * float64(holdTimeInMonths), nil
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
