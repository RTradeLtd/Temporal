package rtfs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
)

// IpfsManager is our helper wrapper for IPFS
type IpfsManager struct {
	Shell           *ipfsapi.Shell
	PubSub          *ipfsapi.PubSubSubscription
	KeystoreManager *KeystoreManager
	KeystoreEnabled bool
	PubTopic        string

	nodeAPIAddr string
}

// Initialize is used ot initialize our Ipfs manager struct
func Initialize(pubTopic, connectionURL string) (*IpfsManager, error) {
	manager := IpfsManager{
		Shell:       EstablishShellWithNode(connectionURL),
		PubTopic:    pubTopic,
		nodeAPIAddr: connectionURL,
	}
	manager.SetTimeout(time.Minute * 1)
	_, err := manager.Shell.ID()
	return &manager, err
}

// EstablishShellWithNode is used to establish our api shell for the ipfs node
func EstablishShellWithNode(url string) *ipfsapi.Shell {
	if url == "" {
		shell := ipfsapi.NewShell("localhost:5001")
		return shell
	}
	shell := ipfsapi.NewShell(url)
	return shell
}

// SetTimeout is used to set a timeout for our api client
func (im *IpfsManager) SetTimeout(time time.Duration) {
	im.Shell.SetTimeout(time)
}

// CreateKeystoreManager is used to create a key store manager for ipfs keys
// for now it just uses a file system keystore manager
func (im *IpfsManager) CreateKeystoreManager() error {
	km, err := GenerateKeystoreManager()
	if err != nil {
		return err
	}
	im.KeystoreManager = km
	im.KeystoreEnabled = true
	return nil
}

// PublishToIPNSDetails is used for fine grained control over IPNS record publishing
func (im *IpfsManager) PublishToIPNSDetails(contentHash, keyName string, lifetime, ttl time.Duration, resolve bool) (*ipfsapi.PublishResponse, error) {
	if !im.KeystoreEnabled {
		return nil, errors.New("attempting to create ipns entry with dynamic keys keystore is not enabled/generated yet")
	}
	keyPresent, err := im.KeystoreManager.CheckIfKeyExists(keyName)
	if err != nil {
		return nil, err
	}
	if !keyPresent {
		return nil, errors.New("attempting to sign with non existent key")
	}
	resp, err := im.Shell.PublishWithDetails(contentHash, keyName, lifetime, ttl, resolve)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Pin is a wrapper method to pin a hash to the local node,
// but also alert the rest of the local nodes to pin
// after which the pin will be sent to the cluster
func (im *IpfsManager) Pin(hash string) error {
	err := im.Shell.Pin(hash)
	if err != nil {
		// TODO: add error reporting
		fmt.Println(err)
		return err
	}
	return nil
}

// Add is a wrapper used to add a file to IPFS
// currently until https://github.com/ipfs/go-ipfs/issues/5376 it is added with no pin
// thus a manual pin must be triggered afterwards
func (im *IpfsManager) Add(r io.Reader) (string, error) {
	hash, err := im.Shell.AddNoPin(r)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// GetObjectFileSizeInBytes is used to retrieve the cumulative byte size of an object
func (im *IpfsManager) GetObjectFileSizeInBytes(key string) (int, error) {
	stat, err := im.Shell.ObjectStat(key)
	if err != nil {
		return 0, err
	}
	return stat.CumulativeSize, nil
}

// ObjectStat is used to retrieve the stats about an object
func (im *IpfsManager) ObjectStat(key string) (*ipfsapi.ObjectStats, error) {
	stat, err := im.Shell.ObjectStat(key)
	if err != nil {
		return nil, err
	}
	return stat, nil
}

// ParseLocalPinsForHash checks whether or not a pin is present
func (im *IpfsManager) ParseLocalPinsForHash(hash string) (bool, error) {
	pins, err := im.Shell.Pins()
	if err != nil {
		return false, err
	}
	info := pins[hash]

	if info.Type != "" {
		return true, nil
	}
	return false, nil
}

// PublishPubSubMessage is used to publish a message to the given topic
func (im *IpfsManager) PublishPubSubMessage(topic string, data string) error {
	fmt.Println("publishing data")
	if topic == "" && data == "" {
		return errors.New("invalid topic and data")
	}
	err := im.Shell.PubSubPublish(topic, data)
	if err != nil {
		fmt.Println("error publishing data ", err.Error())
		return err
	}
	fmt.Println("published data")
	return nil
}

// BuildCustomRequest is used to build a custom request
func (im *IpfsManager) BuildCustomRequest(ctx context.Context, url, commad string, opts map[string]string, args ...string) (*ipfsapi.Response, error) {
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

// DHTFindProvs is used to find providers of a given CID
// Currently bugged and wil only fetch 1 provider
func (im *IpfsManager) DHTFindProvs(cid, numProviders string) error {
	var (
		opts = map[string]string{
			"num-providers": numProviders,
		}
		cmd = "dht/findprovs"
		out = DHTFindProvsResponse{}
	)
	resp, err := im.BuildCustomRequest(context.Background(),
		im.nodeAPIAddr, cmd, opts, cid)
	if err != nil {
		return err
	}
	return resp.Decode(&out)
}
