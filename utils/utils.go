package utils

import (
	"math/big"
	"time"

	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
	"github.com/c2h5oh/datasize"
)

// ConnectionURL is the url used to connect to geth via rpc
var ConnectionURL = "http://192.168.1.245:8545"

// IpcPath is the file path used to connect to geth via ipc
var IpcPath = "/media/solidity/fuck/Rinkeby/datadir/geth.ipc"

// this is a testing parameter for now, exact costs will be detailed at a later time
var usdPerGigabytePerMonth = 0.134

// NilTime is used to compare empty time
var NilTime time.Time

// CalculatePinCost is used to calculate the cost of pining a particular content hash
func CalculatePinCost(contentHash string, holdTimeInMonths int64, shell *ipfsapi.Shell) (float64, error) {
	objectStat, err := shell.ObjectStat(contentHash)
	if err != nil {
		return float64(0), err
	}
	sizeInBytes := objectStat.CumulativeSize
	gigaInBytes := datasize.GB.Bytes()
	gigabytesInt := int64(gigaInBytes)
	sizeInBytesFloat := float64(sizeInBytes)
	gigabytesFloat := float64(gigabytesInt)
	objectSizeInGigabytesFloat := sizeInBytesFloat / gigabytesFloat
	totalCostFloat := objectSizeInGigabytesFloat * float64(holdTimeInMonths)
	return totalCostFloat, nil
}

func CalculateFileCost(holdTimeInMonths, size int64) float64 {
	gigabytesFloat := float64(datasize.GB.Bytes())
	sizeFloat := float64(size)
	sizeGigabytesFloat := sizeFloat / gigabytesFloat
	totalCostUSDFloat := sizeGigabytesFloat * float64(holdTimeInMonths)
	return totalCostUSDFloat
}

func CalculateFileSizeInGigaBytes(size int64) int64 {
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
