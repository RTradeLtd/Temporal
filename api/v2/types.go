package v2

import (
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/kaas"
	xss "github.com/dvwright/xss-mw"

	pbLens "github.com/RTradeLtd/grpc/lensv2"
	pbOrch "github.com/RTradeLtd/grpc/nexus"
	pbSigner "github.com/RTradeLtd/grpc/pay"
)

var (
	xssMdlwr               xss.XssMw
	dev                    = false
	devTermsAndServiceURL  = "https://gateway.temporal.cloud/ipns/docs.dev.ts.temporal.cloud"
	prodTermsAndServiceURL = "https://gateway.temporal.cloud/ipns/docs.ts.temporal.cloud"
	alreadyUploadedMessage = "it seems like you have uploaded content matching this hash already. To save your credits, no charge was placed and the call was gracefully aborted. Please contact support@rtradetechnologies.com if you believe this is an issue"

	whiteListedProxyCalls = map[string]bool{
		// file add
		"add": true,

		// block commands
		"block/get":  true,
		"block/put":  true,
		"block/stat": true,

		// cat commands
		"cat": true,

		// cid commands
		"cid/base32": true,
		"cid/bases":  true,
		"cid/codecs": true,
		"cid/format": true,
		"cid/hashes": true,

		// dag commands
		"dag/get":     true,
		"dag/put":     true,
		"dag/resolve": true,

		// dht comands
		"dht/findpeer":  true,
		"dht/findprovs": true,
		"dht/get":       true,
		// disabling these since
		// we dont want people to be able
		// to randomly put and query the dag
		// as this could be a potentially
		// cpu intensive task
		// "dht/put": true,
		// "dht/query": true,

		// dns commands - alow resolving dns links
		"dns": true,

		"get": true,

		// name commands
		"name/resolve": true,

		// object commands
		"object/data":  true,
		"object/diff":  true,
		"object/get":   true,
		"object/links": true,
		"object/new":   true,
		"object/stat":  true,
		// object/patch will match any patch related commands
		// object/patch needs special handling so its temporarily disabled

		// pin commands
		"pin/add": true,
		// pin/verify allow verification that recursive pins are complete
		// this could potentially be dangerous so we need to
		// think whether or not we should enable thing
		// "pin/verify": true,

		// ping commands
		// this is temporarily disabled sicne it could result in a lot of
		// unneeded spamming
		// "ping": true,

		// pubsub commands
		// disabling ls since it could be used to snoop on other topics
		// pubsub/ls
		// revealing who we are pubsubbing with could be dangerous
		// need to think about whether or not this should be enbaled
		// pubsub/peers
		"pubsub/pub": true,
		// need to think whether or not we want people
		// to be able to subscribe to any topic
		// "pubsub/sub": true,

		// resolve commands
		"resolve": true,

		// swarm commands
		// determine whether or not we want to allow people
		// to connect to random peers. could potentially
		// help with content discover though
		// "swarm/connect": true,

		// tar commands
		// determine whether or not we want to allow people
		// to add tarballs
		// "tar/add": true
		"tar/cat": true,

		// urlstore commands
		// determine if we want ot do this
		// "urlstore/add": true,

		// misc
		"refs": true,
	}
)

// Options is used to non-critical options
type Options struct {
	DebugLogging bool
	DevMode      bool
}

// Clients is used to configure service clients we use
type Clients struct {
	Lens   pbLens.LensV2Client
	Orch   pbOrch.ServiceClient
	Signer pbSigner.SignerClient
}

// CreditRefund is a data object to contain refund information
type CreditRefund struct {
	Username string
	CallType string
	Cost     float64
}

type queues struct {
	pin     *queue.Manager
	cluster *queue.Manager
	email   *queue.Manager
	ipns    *queue.Manager
	key     *queue.Manager
	dash    *queue.Manager
	eth     *queue.Manager
}

// kaas key managers
type keys struct {
	kb1 *kaas.Client
	kb2 *kaas.Client
}
