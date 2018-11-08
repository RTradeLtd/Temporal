package dash

// DuffsToDash is used to convert units of duffs to dash
func DuffsToDash(duffs float64) float64 {
	return duffs / float64(10^8)
}
