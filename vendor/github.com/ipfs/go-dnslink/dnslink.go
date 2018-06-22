/*
Package dnslink implements a dns link resolver. dnslink is a basic
standard for placing traversable links in dns itself. See dnslink.info

A dnslink is a path link in a dns TXT record, like this:

  dnslink=/ipfs/QmR7tiySn6vFHcEjBeZNtYGAFh735PJHfEMdVEycj9jAPy

For example:

  > dig TXT ipfs.io
  ipfs.io.  120   IN  TXT  dnslink=/ipfs/QmR7tiySn6vFHcEjBeZNtYGAFh735PJHfEMdVEycj9jAPy

This package eases resolving and working with thse dns links. For example:

  import (
    dnslink "github.com/jbenet/go-dnslink"
  )

  link, err := dnslink.Resolve("ipfs.io")
  // link = "/ipfs/QmR7tiySn6vFHcEjBeZNtYGAFh735PJHfEMdVEycj9jAPy"

It even supports recursive resolution. Suppose you have three domains with
dnslink records like these:

  > dig TXT foo.com
  foo.com.  120   IN  TXT  dnslink=/dns/bar.com/f/o/o
  > dig TXT bar.com
  bar.com.  120   IN  TXT  dnslink=/dns/long.test.baz.it/b/a/r
  > dig TXT long.test.baz.it
  long.test.baz.it.  120   IN  TXT  dnslink=/b/a/z

Expect these resolutions:

  dnslink.ResolveN("long.test.baz.it", 0) // "/dns/long.test.baz.it"
  dnslink.Resolve("long.test.baz.it")     // "/b/a/z"

  dnslink.ResolveN("bar.com", 1)          // "/dns/long.test.baz.it/b/a/r"
  dnslink.Resolve("bar.com")              // "/b/a/z/b/a/r"

  dnslink.ResolveN("foo.com", 1)          // "/dns/bar.com/f/o/o/"
  dnslink.ResolveN("foo.com", 2)          // "/dns/long.test.baz.it/b/a/r/f/o/o/"
  dnslink.Resolve("foo.com")              // "/b/a/z/b/a/r/f/o/o"

*/
package dnslink

import (
	"errors"
	"net"
	"path"
	"strings"

	isd "github.com/jbenet/go-is-domain"
)

// DefaultDepthLimit controls how many dns links to resolve through before
// returning. Users can override this default.
const DefaultDepthLimit = 16

// MaximumDepthLimit governs the max number of recursive resolutions.
const MaximumDepthLimit = 256

var (
	// ErrInvalidDomain is returned when a string representing a domain name
	// is not actually a valid domain.
	ErrInvalidDomain = errors.New("not a valid domain name")

	// ErrInvalidDnslink is returned when the dnslink entry in a TXT record
	// does not follow the proper dnslink format ("dnslink=<path>")
	ErrInvalidDnslink = errors.New("not a valid dnslink entry")

	// ErrResolveFailed is returned when a resolution failed, most likely
	// due to a network error.
	ErrResolveFailed = errors.New("link resolution failed")

	// ErrResolveLimit is returned when a recursive resolution goes over
	// the limit.
	ErrResolveLimit = errors.New("resolve depth exceeded")
)

// LookupTXTFunc is a function that looks up a TXT record in some dns resovler.
// This is useful for testing or passing your own dns resolution process, which
// could take into account non-standard TLDs like .bit, .onion, .ipfs, etc.
type LookupTXTFunc func(name string) (txt []string, err error)

// Resolve is the simplest way to use this package. It simply resolves the
// dnslink at a particular domain. It will recursively keep resolving until
// reaching the DefaultDepthLimit. If the depth is reached, Resolve will return
// the last value retrieved, and ErrResolveLimit.
// If TXT records are found but are not valid dnslink records, Resolve will
// return ErrInvalidDnslink. Resolve will check every TXT record returned.
// If resolution fails otherwise, Resolve will return ErrResolveFailed
func Resolve(domain string) (string, error) {
	return defaultResolver.Resolve(domain)
}

// ResolveN is just like Resolve, with the option to specify a maximum
// resolution depth.
func ResolveN(domain string, depth int) (string, error) {
	return defaultResolver.ResolveN(domain, depth)
}

// Resolver implements a dnslink Resolver on DNS domains.
// This struct is here for composing dnslink resolution with other
// types of resolvers.
type Resolver struct {
	lookupTXT  LookupTXTFunc
	depthLimit int
	// TODO: maybe some sort of caching?
	// cache would need a timeout
}

// defaultResolver is a resolver used by the main package-level functions.
var defaultResolver = &Resolver{}

// NewResolver constructs a new dnslink resolver. The given defaultDepth
// will be the maximum depth used by the Resolve function.
func NewResolver(defaultDepth int) *Resolver {
	return &Resolver{net.LookupTXT, defaultDepth}
}

func (r *Resolver) setDefaults() {
	// check internal params
	if r.lookupTXT == nil {
		r.lookupTXT = net.LookupTXT
	}
	if r.depthLimit < 1 {
		r.depthLimit = DefaultDepthLimit
	}
	if r.depthLimit > MaximumDepthLimit {
		r.depthLimit = MaximumDepthLimit
	}
}

// Resolve resolves the dnslink at a particular domain. It will recursively
// keep resolving until reaching the defaultDepth of Resolver. If the depth
// is reached, Resolve will return the last value retrieved, and ErrResolveLimit.
// If TXT records are found but are not valid dnslink records, Resolve will
// return ErrInvalidDnslink. Resolve will check every TXT record returned.
// If resolution fails otherwise, Resolve will return ErrResolveFailed
func (r *Resolver) Resolve(domain string) (string, error) {
	return r.ResolveN(domain, DefaultDepthLimit)
}

// ResolveN is just like Resolve, with the option to specify a maximum
// resolution depth.
func (r *Resolver) ResolveN(domain string, depth int) (link string, err error) {
	tail := ""
	for i := 0; i < depth; i++ {
		link, err = r.resolveOnce(domain)
		if err != nil {
			return "", err
		}

		// if does not have /dns/ as a prefix, done.
		if !strings.HasPrefix(link, "/dns/") {
			return link + tail, nil // done
		}

		// keep resolving
		d, rest, err := ParseLinkDomain(link)
		if err != nil {
			return "", err
		}

		domain = d
		tail = rest + tail
	}
	return "/dns/" + domain + tail, ErrResolveLimit
}

// resolveOnce implements resolver.
func (r *Resolver) resolveOnce(domain string) (p string, err error) {
	r.setDefaults()

	if !isd.IsDomain(domain) {
		return "", ErrInvalidDomain
	}

	txt, err := r.lookupTXT(domain)
	if err != nil {
		return "", err
	}

	err = ErrResolveFailed
	for _, t := range txt {
		p, err = ParseTXT(t)
		if err == nil {
			return p, nil
		}
	}

	return "", err
}

// ParseTXT parses a TXT record value for a dnslink value.
// The TXT record must follow the dnslink format:
//   TXT dnslink=<path>
//   TXT dnslink=/foo/bar/baz
// ParseTXT will return ErrInvalidDnslink if parsing fails.
func ParseTXT(txt string) (string, error) {
	parts := strings.SplitN(txt, "=", 2)
	if len(parts) == 2 && parts[0] == "dnslink" && strings.HasPrefix(parts[1], "/") {
		return path.Clean(parts[1]), nil
	}

	return "", ErrInvalidDnslink
}

// ParseLinkDomain parses a domain from a dnslink path.
// The link path must follow the dnslink format:
//   /dns/<domain>/<path>
//   /dns/ipfs.io
//   /dns/ipfs.io/blog/0-hello-worlds
// ParseLinkDomain will return ErrInvalidDnslink if parsing fails,
// and ErrInvalidDomain if the domain is not valid.
func ParseLinkDomain(txt string) (string, string, error) {
	parts := strings.SplitN(txt, "/", 4)
	if len(parts) < 3 || parts[0] != "" || parts[1] != "dns" {
		return "", "", ErrInvalidDnslink
	}

	domain := parts[2]
	if !isd.IsDomain(domain) {
		return "", "", ErrInvalidDomain
	}

	rest := ""
	if len(parts) > 3 {
		rest = "/" + parts[3]
	}
	return domain, rest, nil
}
