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
	devTermsAndServiceURL  = "..."
	prodTermsAndServiceURL = "..."
	alreadyUploadedMessage = "it seems like you have uploaded content matching this hash already. To save your credits, no charge was placed and the call was gracefully aborted. Please contact support@rtradetechnologies.com if you believe this is an issue"
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
