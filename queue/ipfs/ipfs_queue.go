package ipfs

import (
	"fmt"

	"github.com/RTradeLtd/Temporal/api/rtfs_cluster"

	"github.com/RTradeLtd/Temporal/api/rtfs"
)

/*
	This is the IPFS queue which listens for pubsub messages on the pic topic, and then pins that to the local node
*/

// Initialize is used to intiialize the ifps queue
func Initialize() {
	manager := rtfs.Initialize("")
	manager.SubscribeToPubSubTopic(manager.PinTopic)
	pubSub := manager.PubSub
	clusterManager := rtfs_cluster.Initialize()
	for {
		subRecord, err := pubSub.Next()
		if err != nil {
			fmt.Println("erorr detected")
			fmt.Println(err)
			continue
		}
		dataString := string(subRecord.Data())
		fmt.Printf("Pinning %s to cluster\n", dataString)
		decodedDataString := clusterManager.DecodeHashString(dataString)
		err = clusterManager.Pin(decodedDataString)
		if err != nil {
			fmt.Println("error detected")
			fmt.Println(err)
			continue
		}
	}
}
