package v2

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/RTradeLtd/Temporal/eh"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

func (api *API) proxyIPFS(c *gin.Context) {
	if !dev {
		Fail(c, errors.New("reverse proxy ipfs api is not available in prod"))
		return
	}
	username, err := GetAuthenticatedUserFromContext(c)
	if err != nil {
		api.LogError(c, err, eh.NoAPITokenError)(http.StatusBadRequest)
		return
	}
	if err := checkCall(c.Param("ipfs")); err != nil {
		api.LogError(c, err, err.Error())(http.StatusBadRequest)
		return
	}
	var protocol string
	if c.Request.URL.Scheme != "" {
		protocol = c.Request.URL.Scheme
	} else {
		protocol = "http://"
	}
	var (
		address = api.cfg.IPFS.APIConnection.Host + ":" + api.cfg.IPFS.APIConnection.Port
		target  = fmt.Sprintf("%s%s%s", protocol, address, c.Request.RequestURI)
	)
	remote, err := url.Parse(target)
	if err != nil {
		api.LogError(c, err, err.Error())(http.StatusInternalServerError)
		return
	}
	logger := api.l.Named("proxy").With("user", username)
	// use remote.Query() to get the url values so we can parse calls
	// in the case of pin/add, the hash being pinned is under "args"
	// todo: perform deeper validation of calls, ensuring we properly update
	// the database, and handle stuff like credits, invalid balances, err..
	newProxy(
		remote,
		logger,
		false,
	).ServeHTTP(c.Writer, c.Request)
	//proxy.ServeHTTP(c.Writer, c.Request)
	api.l.Info("reverse proxy request served", "user", username)
}

func newProxy(target *url.URL, l *zap.SugaredLogger, direct bool) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// if set up as an indirect proxy, we need to remove delgator-specific
			// leading elements, e.g. /networks/test_network/api, from the path and
			// accomodate for specific cases
			if !direct {
				req.URL.Path = "/api" + stripLeadingSegments(req.URL.Path)
			}

			// set other URL properties
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			l.Infow(
				"forwarding request",
				"url", req.URL.String(),
				"path", req.URL.Path,
			)
		},
	}
}

func stripLeadingSegments(path string) string {
	var expected = 5
	var parts = strings.SplitN(path, "/", expected)
	if len(parts) == expected {
		return "/" + parts[expected-1]
	}
	return path
}

// checkCall is used to prevent certain calls
// from being proxied through. This is limited
// to calls that involve removing data, listing pinned data
// creating keys, publishing, etc...
func checkCall(request string) error {
	trimmed := strings.TrimPrefix(request, "/api/v0/")
	if !whiteListedProxyCalls[trimmed] {
		return errors.New("sorry the API call you are using is not white listed for reverse proxying")
	}
	return nil
}
