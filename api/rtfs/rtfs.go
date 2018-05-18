package rtfs

import (
	"errors"
	"fmt"

	"github.com/RTradeLtd/Temporal/api/rtfs_cluster"
	"github.com/RTradeLtd/Temporal/models"

	ipfsapi "github.com/ipfs/go-ipfs-api"
)

type IpfsManager struct {
	Shell    *ipfsapi.Shell
	PubSub   *ipfsapi.PubSubSubscription
	PinTopic string
}

func Initialize(pinTopic string) *IpfsManager {
	if pinTopic == "" {
		pinTopic = "ipfs-pins"
	}
	manager := IpfsManager{}
	manager.Shell = establishShellWithNode("")
	manager.PinTopic = pinTopic
	return &manager
}

// Pin is a wrapper method to pin a hash to the local node,
// but also alert the rest of the local nodes to pin
// after which the pin will be sent to the cluster
func (im *IpfsManager) Pin(hash string) error {
	err := im.Shell.Pin(hash)
	if err != nil {
		// TODO: add error reporting
		return err
	}
	im.PublishPubSubMessage(im.PinTopic, hash)
	cm := rtfs_cluster.Initialize()
	decoded := cm.DecodeHashString(hash)
	err = cm.Pin(decoded)
	if err != nil {
		return err
	}
	return nil
}

// Add is a wrapper used to add a file to the node
func (im *IpfsManager) Add(file string) error {
	return nil
}

func establishShellWithNode(url string) *ipfsapi.Shell {
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
