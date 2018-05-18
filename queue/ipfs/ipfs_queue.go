package ipfs

import (
	"github.com/RTradeLtd/Temporal/api/rtfs"
)

/*
	This is the IPFS queue which listens for pubsub messages on the pic topic, and then pins that to the local node
*/

// Initialize is used to intiialize the ifps queue
func Initialize() {
	manager := rtfs.Initialize("")
	manager.SubscribeToPubSubTopic(manager.PinTopic)
	manager.ConsumeSubscriptionToPin(manager.PubSub)
}
