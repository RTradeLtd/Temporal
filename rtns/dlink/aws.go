package dlink

/*
This provides access to create dnslink TXT records on AWS Route53 Domains
*/

import (
	"errors"
	"fmt"

	"github.com/mitchellh/goamz/aws"
	r "github.com/mitchellh/goamz/route53"
	route53 "github.com/segmentio/go-route53"
)

// AwsLinkManager is used to manage a connection to aws
type AwsLinkManager struct {
	Auth   aws.Auth
	Region aws.Region
	Client *route53.Client
	Zone   string
}

// GenerateAwsLinkManager is used to generate the configs needed to interact with Route53
// and create dnslink TXT records, to allow for the resolution of human readable names to
// IPNS entries
func GenerateAwsLinkManager(authMethod, accessKey, secretKey, zone string, region aws.Region) (*AwsLinkManager, error) {
	var alm AwsLinkManager
	var auth aws.Auth
	var err error
	if zone == "" {
		return nil, errors.New("zone is empty")
	}
	switch authMethod {
	case "env":
		auth, err = aws.EnvAuth()
		if err != nil {
			return nil, err
		}
	case "get":
		if accessKey == "" {
			return nil, errors.New("accessKey is empty")
		}
		if secretKey == "" {
			return nil, errors.New("secretKey is empty")
		}
		auth, err = aws.GetAuth(accessKey, secretKey)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid authMethod provided")
	}
	dns := route53.New(auth, region)
	alm.Auth = auth
	alm.Region = region
	alm.Client = dns
	alm.Zone = zone
	return &alm, nil
}

// AddDNSLinkEntry is used to a dnslink record
func (alm *AwsLinkManager) AddDNSLinkEntry(name, value string) (*r.ChangeResourceRecordSetsResponse, error) {
	// TODO: add a filter to prevent NSFW words
	resp, err := alm.Client.Zone(alm.Zone).Add("TXT", name, value)
	if err != nil {
		return nil, err
	}

	fmt.Println(resp.ChangeInfo)
	return nil, nil
}
