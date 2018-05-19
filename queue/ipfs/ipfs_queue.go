package ipfs

import (
	"errors"
	"fmt"

	"github.com/RTradeLtd/Temporal/api/rtfs_cluster"

	"github.com/RTradeLtd/Temporal/api/rtfs"
	ipfsapi "github.com/ipfs/go-ipfs-api"
)

// PubSubQueueManager is a helper struct used to create a queue system that uses ipfs pubsub
type PubSubQueueManager struct {
	Shell      *ipfsapi.Shell
	PubSub     *ipfsapi.PubSubSubscription
	QueueTopic string
}

// Initialize is used to intiialize the ifps queue smanager
func Initialize(topic string) (*PubSubQueueManager, error) {
	psm := PubSubQueueManager{
		QueueTopic: topic,
	}
	shell := rtfs.EstablishShellWithNode("")
	psm.Shell = shell
	online := psm.Shell.IsUp()
	if !online {
		return nil, errors.New("node is not online")
	}
	subscription, err := psm.Shell.PubSubSubscribe(psm.QueueTopic)
	if err != nil {
		return nil, err
	}
	psm.PubSub = subscription
	return &psm, nil
}

// ParseClusterPinTopic is used to parse through any pubsub cluster pin topics, and pin the contained data
func (pm *PubSubQueueManager) ParseClusterPinTopic() {
	fmt.Println("test")
	if pm.QueueTopic != rtfs.ClusterPubSubTopic {
		fmt.Println("pubsub queue manager is not configured to listen and parse cluster pins")
		return
	}
	cm := rtfs_cluster.Initialize()
	for {
		fmt.Println("test 1")
		record, err := pm.PubSub.Next()
		if err != nil {
			fmt.Println(err)
			// todo: add error handling
			continue
		}
		fmt.Println("record detected")
		dataString := string(record.Data())
		decodedDataString := cm.DecodeHashString(dataString)
		err = cm.Pin(decodedDataString)
		if err != nil {
			// todo: add errror handling
			fmt.Println("error pinning hash to cluster")
			fmt.Println(err)
			continue
		}
		fmt.Println("pin succeeded")
	}
}

// ListenToClusterPinTopic is used to subscribe, and listen for events indicating
// a pin needs to be added to the cluster
func ListenToClusterPinTopic() {
	manager := rtfs.Initialize(rtfs.ClusterPubSubTopic)
	manager.SubscribeToPubSubTopic(manager.PubTopic)
	pubSub := manager.PubSub
	clusterManager := rtfs_cluster.Initialize()
	for {
		subRecord, err := pubSub.Next()
		if err != nil {
			fmt.Println("error detected")
			fmt.Println(err)
			continue
		}
		dataString := string(subRecord.Data())
		decodedDataString := clusterManager.DecodeHashString(dataString)
		err = clusterManager.Pin(decodedDataString)
		if err != nil {
			fmt.Println("error detected")
			fmt.Println(err)
			continue
		}
		fmt.Println("pinned successfully")
	}
}
