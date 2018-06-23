package rtfs

import (
	"errors"
	"fmt"

	ipfsapi "github.com/RTradeLtd/go-ipfs-api"
)

var ClusterPubSubTopic = "ipfs-cluster"

type IpfsManager struct {
	Shell           *ipfsapi.Shell
	PubSub          *ipfsapi.PubSubSubscription
	KeystoreManager *KeystoreManager
	KeystoreEnabled bool
	PubTopic        string
}

func Initialize(pubTopic, connectionURL string) (*IpfsManager, error) {
	if pubTopic == "" {
		pubTopic = ClusterPubSubTopic
	}
	manager := IpfsManager{}
	manager.Shell = EstablishShellWithNode(connectionURL)
	_, err := manager.Shell.ID()
	if err != nil {
		return nil, err
	}
	manager.PubTopic = pubTopic
	return &manager, nil
}

func EstablishShellWithNode(url string) *ipfsapi.Shell {
	if url == "" {
		shell := ipfsapi.NewShell("localhost:5001")
		return shell
	}
	shell := ipfsapi.NewShell(url)
	return shell
}

func (im *IpfsManager) CreateKeystoreManager() error {
	km, err := GenerateKeystoreManager()
	if err != nil {
		return err
	}
	im.KeystoreManager = km
	im.KeystoreEnabled = true
	return nil
}

// TODO: lock this down upstream. We will need to make sure that they own the key they are attempting to access
func (im *IpfsManager) PublishToIPNSDetails(contentHash string, lifetime string, ttl string, keyName, keyID string, resolve bool) (*ipfsapi.PublishResponse, error) {
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
	resp, err := im.Shell.PublishWithDetails(contentHash, lifetime, ttl, keyID, resolve)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Pin is a wrapper method to pin a hash to the local node,
// but also alert the rest of the local nodes to pin
// after which the pin will be sent to the cluster
func (im *IpfsManager) Pin(hash string) error {
	fmt.Println("pinning hash locally")
	err := im.Shell.Pin(hash)
	fmt.Println("hash pinned locally")
	if err != nil {
		// TODO: add error reporting
		fmt.Println(err)
		return err
	}
	return nil
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

// SubscribeToPubSubTopic is used to subscribe to a pubsub topic
func (im *IpfsManager) SubscribeToPubSubTopic(topic string) error {
	if topic == "" {
		return errors.New("invalid topic name")
	}
	// create a pubsub subscription according to topic name
	subscription, err := im.Shell.PubSubSubscribe(topic)
	if err != nil {
		return err
	}
	// store the pubsub scription
	im.PubSub = subscription
	return nil
}

// ConsumeSubscription is used to consume a pubsub subscription
func (im *IpfsManager) ConsumeSubscription(sub *ipfsapi.PubSubSubscription) error {
	for {
		// we should try and add some logic to unmarshal this data instead
		subRecord, err := sub.Next()
		if err != nil {
			continue
		}
		fmt.Println(string(subRecord.Data()))
	}
}

// ConsumeSubscriptionToPin is used to consume a pubsub subscription, with the intent of pinning the content referred to on the local ipfs node
func (im *IpfsManager) ConsumeSubscriptionToPin(sub *ipfsapi.PubSubSubscription) error {
	for {
		// we should try and add some logic to unmarshal this data instead
		subRecord, err := sub.Next()
		if err != nil {
			continue
		}
		cidString := string(subRecord.Data())
		fmt.Println("pinning ", cidString)
		err = im.Shell.Pin(cidString)
		if err != nil {
			fmt.Println("pin failed")
			continue
		}
		fmt.Println("pin succeeded")
	}
}

// PublishPubSubMessage is used to publish a message to the given topic
func (im *IpfsManager) PublishPubSubMessage(topic string, data string) error {
	if topic == "" && data == "" {
		return errors.New("invalid topic and data")
	}
	err := im.Shell.PubSubPublish(topic, data)
	if err != nil {
		return err
	}
	return nil
}
