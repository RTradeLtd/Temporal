package dlink

/*
This provides access to create dnslink TXT records on AWS Route53 Domains
*/

import (
	"errors"

	"github.com/mitchellh/goamz/aws"
)

type AwsLinkManager struct {
	AwsAuth aws.Auth
}

func GenerateAwsLinkManager(authMethod, accessKey, secretKey string) (*AwsLinkManager, error) {
	var alm AwsLinkManager
	var auth aws.Auth
	var err error
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
	alm.AwsAuth = auth
	return &alm, nil
}
