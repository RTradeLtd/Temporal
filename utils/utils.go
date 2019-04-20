package utils

import (
	"math/big"

	"github.com/RTradeLtd/database/v2/models"
	"github.com/RTradeLtd/rtfs/v2"
	"github.com/c2h5oh/datasize"
)

// CalculatePinCost is used to calculate the cost of pining a particular content hash
func CalculatePinCost(username, contentHash string, holdTimeInMonths int64, im rtfs.Manager, um *models.UsageManager) (float64, error) {
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
	// get the users usage model
	usage, err := um.FindByUserName(username)
	if err != nil {
		return 0, err
	}
	// if they are free tier, they don't incur data charges
	if usage.Tier == models.Free {
		return 0, nil
	}
	// dynamic pricing based on their usage tier
	costPerMonthFloat := objectSizeInGigabytesFloat * usage.Tier.PricePerGB()
	return costPerMonthFloat * float64(holdTimeInMonths), nil
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
	if usage.Tier == models.Free {
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
