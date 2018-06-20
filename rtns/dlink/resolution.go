package dlink

import (
	"github.com/ipfs/go-dnslink"
)

/*
This is used to handle resolution of dnslink records
*/

// ResolveURL is used to resolve a DNS link text record
func ResolveURL(url string) (string, error) {
	link, err := dnslink.Resolve(url)
	if err != nil {
		return "", err
	}
	return link, nil
}
