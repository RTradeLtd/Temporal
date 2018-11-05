package utils

import (
	"context"

	"google.golang.org/grpc/credentials"
)

// Credentials holds per-rpc metadata for the gRPC clients
type Credentials struct {
	token  string
	secure bool
}

// NewCredentials instantiates a new credentials container
func NewCredentials(token string, withTransportSecurity bool) credentials.PerRPCCredentials {
	return Credentials{token, withTransportSecurity}
}

// GetRequestMetadata retrieves relevant metadata
func (c Credentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": c.token,
	}, nil
}

// RequireTransportSecurity indicates that transport security is required
func (c Credentials) RequireTransportSecurity() bool { return c.secure }
