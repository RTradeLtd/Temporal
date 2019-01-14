package rtfs

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strings"
	"time"

	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
)

// IpfsManager is our helper wrapper for IPFS
type IpfsManager struct {
	shell       *ipfsapi.Shell
	nodeAPIAddr string
}

// NewManager is used to initialize our Ipfs manager struct
func NewManager(ipfsURL string, timeout time.Duration) (*IpfsManager, error) {
	// set up shell
	sh := newShell(ipfsURL)
	sh.SetTimeout(time.Minute * 5)
	if _, err := sh.ID(); err != nil {
		return nil, fmt.Errorf("failed to connect to ipfs node at '%s': %s", ipfsURL, err.Error())
	}

	// instantiate manager
	return &IpfsManager{
		shell:       sh,
		nodeAPIAddr: ipfsURL,
	}, nil
}

// NodeAddress returns the node the manager is connected to
func (im *IpfsManager) NodeAddress() string { return im.nodeAPIAddr }

// Add is a wrapper used to add a file to IPFS
func (im *IpfsManager) Add(r io.Reader) (string, error) {
	return im.shell.Add(r)
}

// DagPut is used to store data as an ipld object
func (im *IpfsManager) DagPut(data interface{}, encoding, kind string) (string, error) {
	return im.shell.DagPut(data, encoding, kind)
}

// DagGet is used to get an ipld object
func (im *IpfsManager) DagGet(cid string, out interface{}) error {
	return im.shell.DagGet(cid, out)
}

// Cat is used to get cat an ipfs object
func (im *IpfsManager) Cat(cid string) ([]byte, error) {
	r, err := im.shell.Cat(cid)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

// Stat is used to retrieve the stats about an object
func (im *IpfsManager) Stat(hash string) (*ipfsapi.ObjectStats, error) {
	return im.shell.ObjectStat(hash)
}

// PatchLink is used to link two objects together
// path really means the name of the link
// create is used to specify whether intermediary nodes should be generated
func (im *IpfsManager) PatchLink(root, path, childHash string, create bool) (string, error) {
	return im.shell.PatchLink(root, path, childHash, create)
}

// AppendData is used to modify the raw data within an object, to a max of 1MB
// Anything larger than 1MB will not be respected by the rest of the network
func (im *IpfsManager) AppendData(root string, data interface{}) (string, error) {
	return im.shell.PatchData(root, false, data)
}

// SetData is used to set the data field of an ipfs object
func (im *IpfsManager) SetData(root string, data interface{}) (string, error) {
	return im.shell.PatchData(root, true, data)
}

// NewObject is used to create a generic object from a template type
func (im *IpfsManager) NewObject(template string) (string, error) {
	return im.shell.NewObject(template)
}

// Pin is a wrapper method to pin a hash.
// pinning prevents GC and persistently stores on disk
func (im *IpfsManager) Pin(hash string) error {
	if err := im.shell.Pin(hash); err != nil {
		return fmt.Errorf("failed to pin '%s': %s", hash, err.Error())
	}
	return nil
}

// CheckPin checks whether or not a pin is present
func (im *IpfsManager) CheckPin(hash string) (bool, error) {
	pins, err := im.shell.Pins()
	if err != nil {
		return false, err
	}
	if info := pins[hash]; info.Type != "" {
		return true, nil
	}
	return false, nil
}

// Publish is used for fine grained control over IPNS record publishing
func (im *IpfsManager) Publish(contentHash, keyName string, lifetime, ttl time.Duration, resolve bool) (*ipfsapi.PublishResponse, error) {
	return im.shell.PublishWithDetails(contentHash, keyName, lifetime, ttl, resolve)
}

// Resolve is used to resolve an IPNS hash
func (im *IpfsManager) Resolve(hash string) (string, error) {
	return im.shell.Resolve(hash)
}

// PubSubPublish is used to publish a a message to the given topic
func (im *IpfsManager) PubSubPublish(topic string, data string) error {
	if topic == "" {
		return errors.New("topic is empty")
	} else if data == "" {
		return errors.New("data is empty")
	}
	return im.shell.PubSubPublish(topic, data)
}

// CustomRequest is used to make a custom request
func (im *IpfsManager) CustomRequest(ctx context.Context, url, commad string,
	opts map[string]string, args ...string) (*ipfsapi.Response, error) {
	req := ipfsapi.NewRequest(ctx, url, commad, args...)
	if len(opts) > 0 {
		currentOpts := req.Opts
		for k, v := range opts {
			currentOpts[k] = v
		}
		req.Opts = currentOpts
	}
	hc := &http.Client{Timeout: time.Minute * 1}
	resp, err := req.Send(hc)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// SwarmConnect is use to open a connection a one or more ipfs nodes
func (im *IpfsManager) SwarmConnect(ctx context.Context, addrs ...string) error {
	return im.shell.SwarmConnect(ctx, addrs...)
}

// DedupAndCalculatePinSize is used to remove duplicate refers to objects for a more accurate pin size cost
// it returns the size of all refs, as well as all unique references
func (im *IpfsManager) DedupAndCalculatePinSize(hash string) (int64, []string, error) {
	// format a multiaddr api to connect to
	parsedIP := strings.Split(im.nodeAPIAddr, ":")
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
