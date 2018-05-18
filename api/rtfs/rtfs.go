package rtfs

import (
	"errors"
	"fmt"

	ipfsapi "github.com/ipfs/go-ipfs-api"
)

type IpfsManager struct {
	Shell  *ipfsapi.Shell
	PubSub *ipfsapi.PubSubSubscription
}

func Initialize() *IpfsManager {
	manager := IpfsManager{}
	manager.Shell = establishShellWithNode("")
	return &manager
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
	subscription, err := im.Shell.PubSubSubscribe(topic)
	if err != nil {
		return err
	}
	im.PubSub = subscription
	return nil
}

func (im *IpfsManager) ConsumeSubscription(sub *ipfsapi.PubSubSubscription) error {
	count := 0
	for {
		if count == 1000 {
			break
		}
		subRecord, err := sub.Next()
		if err != nil {
			return err
		}
		if subRecord == nil {
			continue
		}
		count++
		fmt.Println(subRecord)
	}
	return nil
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
