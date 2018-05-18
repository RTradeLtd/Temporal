package ipfs

import (
	"fmt"

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
	for {
		subRecord, err := pubSub.Next()
		if err != nil {
			fmt.Println("erorr detected")
			fmt.Println(err)
			continue
		}
		dataString := string(subRecord.Data())
		fmt.Println(dataString)
	}
}
