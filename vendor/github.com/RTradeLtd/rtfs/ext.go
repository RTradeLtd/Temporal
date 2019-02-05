package rtfs

// non-class functions

// DedupAndCalculatePinSize is used to remove duplicate refers to objects for a more accurate pin size cost
// it returns the size of all refs, as well as all unique references
func DedupAndCalculatePinSize(hash string, im Manager) (int64, []string, error) {
	// since we  are looking to calculate deduplicated costs,
	// we only want to consider hashes once, ie if they are linked multple times
	// ignore them
	refs, err := im.Refs(hash, true, true)
	if err != nil {
		return 0, nil, err
	}
	// the total size of all data in all references
	var totalDataSize int
	// parse through all references
	for _, ref := range refs {
		// grab object stats for the reference
		refStats, err := im.Stat(ref)
		if err != nil {
			return 0, nil, err
		}
		// update totalDataSize
		totalDataSize = totalDataSize + refStats.DataSize
	}
	return int64(totalDataSize), refs, nil
}
