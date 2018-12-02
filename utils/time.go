package utils

import "time"

// CalculateGarbageCollectDate is used to calculate the date at which data is removed from our system
func CalculateGarbageCollectDate(holdTimeInMonths int) time.Time {
	return time.Now().AddDate(0, holdTimeInMonths, 0)
}
