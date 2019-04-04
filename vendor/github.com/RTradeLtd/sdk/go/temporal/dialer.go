package temporal

import (
	"context"
	"time"

	"github.com/RTradeLtd/sdk/go/temporal/auth"
)

// Credentials holds per-rpc metadata for the gRPC clients
type Credentials struct {
	token    *auth.Token
	secure   bool
	connOpts ConnOpts
}

// NewCredentials instantiates a new credentials container
func NewCredentials(
	token *auth.Token,
	withTransportSecurity bool,
	connOpts ConnOpts,
) *Credentials {
	return &Credentials{token, withTransportSecurity, connOpts}
}

// GetRequestMetadata retrieves relevant metadata. It also refreshes the token
// when appropriate.
func (c *Credentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	if c.token.GetExpire() > (time.Now().Unix() - int64(5*time.Minute)) {
		var err error
		conn, err := Connect(ctx, c.token, c.connOpts)
		if err != nil {
			return nil, err
		}
		c.token, err = NewAuthClient(conn).Refresh(ctx, &auth.Empty{})
		if err != nil {
			return nil, err
		}
	}
	return map[string]string{
		string(MetaKeyAuthorization): "Bearer " + c.token.GetToken(),
	}, nil
}

// RequireTransportSecurity indicates that transport security is required
func (c *Credentials) RequireTransportSecurity() bool { return c.secure }
