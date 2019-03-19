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
	if err := checkCall(c.Param("proxy")); err != nil {
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
	whiteListedCalls := []string{
		// pin commands
		"pin/add",
		// need to consider security issues with pin/update
		//"pin/update",

		// add file
		"add",

		// resolve content
		"resolve",

		// dns comands
		"dns",

		// object commands
		"object/data",
		"object/diff",
		"object/links",
		"object/get",
		// object/put

		// dag commands
		"dag/get",
		"dag/resolve",

		"get",

		"refs",

		// dht comands
		"dht/query",
		"dht/findpeer",
		"dht/findprovs",
		"dht/num-providers",
		"dht/get",
		// dht/put

		// namesys commands
		"name/resolve",

		// block commands
		"block/stat",
		"block/get",
		// "block/put"

		// tar commands
		"tar/cat",
		// "tar/add"
	}
	trimmed := strings.TrimPrefix(request, "/api/v0/")
	var isWhiteListed bool
	for _, v := range whiteListedCalls {
		if trimmed == v {
			isWhiteListed = true
			break
		}
	}
	if !isWhiteListed {
		return errors.New("sorry the API call you are using is not white listed for reverse proxying")
	}
	return nil
}
