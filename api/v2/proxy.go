package v2

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func (api *API) proxyIPFS(c *gin.Context) {
	if err := checkCall(c.Param("ipfs")); err != nil {
		Fail(c, err)
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
		api.LogError(c, err, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(remote.Query())
	proxy := newProxy(remote, false)
	proxy.ServeHTTP(c.Writer, c.Request)
}

func newProxy(target *url.URL, direct bool) *httputil.ReverseProxy {
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
	whiteListedCalls := map[string]bool{
		// pin calls
		"pin/add": true,

		// file add
		"add": true,

		// ipns resolution
		"resolve": true,

		// dns commands
		"dns": true,

		// name commands
		"name/resolve": true,

		// object commands
		"object/data":  true,
		"object/diff":  true,
		"object/links": true,
		"object/get":   true,
		// object/put (this needs special handling)

		// dag commands
		"dag/get":     true,
		"dag/resolve": true,
		// dag/put (this needs special handling)

		// dht comands
		"dht/query":         true,
		"dht/findpeer":      true,
		"dht/findprovs":     true,
		"dht/num-providers": true,
		"dht/get":           true,
		// dht/put (this needs special handling)

		// tar commands
		"tar/cat": true,
		// tar/add (this needs special handling)

		// misc
		"get":  true,
		"refs": true,
	}
	trimmed := strings.TrimPrefix(request, "/api/v0/")
	if !whiteListedCalls[trimmed] {
		return errors.New("sorry the API call you are using is not white listed for reverse proxying")
	}
	return nil
}
