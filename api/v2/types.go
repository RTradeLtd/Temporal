package v2

import (
	clients "github.com/RTradeLtd/Temporal/grpc-clients"
	"github.com/RTradeLtd/Temporal/queue"
	"github.com/RTradeLtd/Temporal/rtfscluster"
	"github.com/RTradeLtd/config"
	"github.com/RTradeLtd/database"
	"github.com/RTradeLtd/database/models"
	pbLens "github.com/RTradeLtd/grpc/lens"
	pbOrch "github.com/RTradeLtd/grpc/nexus"
	pbSigner "github.com/RTradeLtd/grpc/pay"
	"github.com/RTradeLtd/kaas"
	"github.com/RTradeLtd/rtfs"
	xss "github.com/dvwright/xss-mw"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	xssMdlwr               xss.XssMw
	dev                    = false
	devTermsAndServiceURL  = "..."
	prodTermsAndServiceURL = "..."
)

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
