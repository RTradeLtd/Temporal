package rtfs

import (
	"errors"
	"fmt"

	"github.com/RTradeLtd/Temporal/models"

	ipfsapi "github.com/ipfs/go-ipfs-api"
)

var ClusterPubSubTopic = "ipfs-cluster"

type IpfsManager struct {
	Shell    *ipfsapi.Shell
	PubSub   *ipfsapi.PubSubSubscription
	PubTopic string
}

func Initialize(pubTopic string) *IpfsManager {
	if pubTopic == "" {
		pubTopic = ClusterPubSubTopic
	}
	manager := IpfsManager{}
	manager.Shell = EstablishShellWithNode("")
	manager.PubTopic = pubTopic
	return &manager
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

func EstablishShellWithNode(url string) *ipfsapi.Shell {
	if url == "" {
		shell := ipfsapi.NewLocalShell()
		return shell
	}
	shell := ipfsapi.NewShell(url)
	return shell
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

func (im *IpfsManager) PublishPubSubTest(topic string) error {
	upload := models.Upload{
		Hash:          "hash",
		UploadAddress: "address",
	}

	err := im.Shell.PubSubPublish(topic, fmt.Sprint(upload))
	if err != nil {
		return err
	}
	return nil
}
