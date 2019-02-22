package rtfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
)

// IpfsManager is our helper wrapper for IPFS
type IpfsManager struct {
	shell       *ipfsapi.Shell
	nodeAPIAddr string
}

// NewManager is used to instantiate IpfsManager with a connection to an ipfs api.
// if token is provided, we use it to establish an authentication, direct connection
// to an ipfs node api, which involves skipping multiaddr parsing. This is useful
// in situations such as interacting with Nexus' delegator to talk with private ipfs
// networks which use non-standard connection methods.
func NewManager(ipfsURL, token string, timeout time.Duration) (*IpfsManager, error) {
	var sh *ipfsapi.Shell
	if token != "" {
		sh = ipfsapi.NewDirectShell(ipfsURL).WithAuthorization(token)
	} else {
		sh = ipfsapi.NewShell(ipfsURL)
	}
	// validate we have an active connection
	if _, err := sh.ID(); err != nil {
		return nil, fmt.Errorf("failed to connect to ipfs node at '%s': %s", ipfsURL, err.Error())
	}
	// set timeout
	sh.SetTimeout(timeout)
	// instantiate and return manager
	return &IpfsManager{
		shell:       sh,
		nodeAPIAddr: ipfsURL,
	}, nil
}

// NodeAddress returns the node the manager is connected to
func (im *IpfsManager) NodeAddress() string { return im.nodeAPIAddr }

// Add is a wrapper used to add a file to IPFS
func (im *IpfsManager) Add(r io.Reader, options ...ipfsapi.AddOpts) (string, error) {
	return im.shell.Add(r, options...)
}

// AddDir is used to add a directory to ipfs
func (im *IpfsManager) AddDir(dir string) (string, error) {
	return im.shell.AddDir(dir)
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
	var (
		r   io.ReadCloser
		err error
	)
	r, err = im.shell.Cat(cid)
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
	return im.shell.Pin(hash)
}

// PinUpdate is used to update one pin to another, while making sure all objects
// in the new pin are local, followed by removing the old pin.
//
// This is an optimized version of pinning the new content, and then removing the
// old content.
//
// returns the new pin path
func (im *IpfsManager) PinUpdate(from, to string) (string, error) {
	out, err := im.shell.PinUpdate(from, to)
	if err != nil {
		return "", err
	}
	if len(out) == 0 || len(out["Pins"]) == 0 {
		return "", errors.New("failed to retrieve new pin paths")
	}
	return out["Pins"][1], nil
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

// Refs is used to retrieve references of a hash
func (im *IpfsManager) Refs(hash string, recursive, unique bool) ([]string, error) {
	refs, err := im.shell.Refs(hash, recursive, unique)
	if err != nil {
		return nil, err
	}
	var references []string
	for ref := range refs {
		references = append(references, ref)
	}
	return references, nil
}
