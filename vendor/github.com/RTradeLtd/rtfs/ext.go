package rtfs

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

// non-class functions

// DedupAndCalculatePinSize is used to remove duplicate refers to objects for a more accurate pin size cost
// it returns the size of all refs, as well as all unique references
func DedupAndCalculatePinSize(hash string, im Manager) (int64, []string, error) {
	// format a multiaddr api to connect to
	parsedIP := strings.Split(im.NodeAddress(), ":")
	multiAddrIP := fmt.Sprintf("/ip4/%s/tcp/%s", parsedIP[0], parsedIP[1])
	// Shell::Refs doesn't seem to return more than 1 hash, and doesn't allow usage of flags like `--unique`
	// will open up a PR with main `go-ipfs-api` to address this, but in the mean time this is a good monkey-patch
	outBytes, err := exec.Command("ipfs", fmt.Sprintf("--api=%s", multiAddrIP), "refs", "--recursive", "--unique", hash).Output()
	if err != nil {
		return 0, nil, err
	}
	// convert exec output to scanner
	scanner := bufio.NewScanner(strings.NewReader(string(outBytes)))
	var refsArray []string
	// iterate over output grabbing hashes
	for scanner.Scan() {
		refsArray = append(refsArray, scanner.Text())
	}
	if scanner.Err() != nil {
		return 0, nil, scanner.Err()
	}
	// the total size of all data in all references
	var totalDataSize int
	// parse through all references
	for _, ref := range refsArray {
		// grab object stats for the reference
		refStats, err := im.Stat(ref)
		if err != nil {
			return 0, nil, err
		}
		// update totalDataSize
		totalDataSize = totalDataSize + refStats.DataSize
	}
	return int64(totalDataSize), refsArray, nil
}
