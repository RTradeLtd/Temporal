package v2

import (
	"github.com/RTradeLtd/ChainRider-Go/dash"
	clients "github.com/RTradeLtd/Temporal/grpc-clients"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfscluster"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
	"github.com/RTradeLtd/rtfs"
	xss "github.com/dvwright/xss-mw"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	pbLens "github.com/RTradeLtd/grpc/lens"
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

// API is our API service
type API struct {
	ipfs        rtfs.Manager
	ipfsCluster *rtfscluster.ClusterManager
	keys        keys
	r           *gin.Engine
	cfg         *config.TemporalConfig
	dbm         *database.Manager
	um          *models.UserManager
	im          *models.IpnsManager
	pm          *models.PaymentManager
	ue          *models.EncryptedUploadManager
	upm         *models.UploadManager
	zm          *models.ZoneManager
	rm          *models.RecordManager
	nm          *models.IPFSNetworkManager
	usage       *models.UsageManager
	l           *zap.SugaredLogger
	signer      pbSigner.SignerClient
	orch        pbOrch.ServiceClient
	lens        pbLens.IndexerAPIClient
	dc          *dash.Client
	queues      queues
	service     string

	version string
}

// Options is used to non-critical options
type Options struct {
	DebugLogging bool
	DevMode      bool
}

// Clients is used to configure service clients we use
type Clients struct {
	Lens   pbLens.IndexerAPIClient
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
	kb1 *clients.KaasClient
	kb2 *clients.KaasClient
}
