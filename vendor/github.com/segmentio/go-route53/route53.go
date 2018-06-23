//
// High-level route53 client.
//
package route53

import r "github.com/mitchellh/goamz/route53"
import "github.com/mitchellh/goamz/aws"

// Client.
type Client struct {
	*r.Route53
}

// New client with the given auth and region.
func New(auth aws.Auth, region aws.Region) *Client {
	return &Client{
		r.New(auth, region),
	}
}

// Zone for `id`.
func (c *Client) Zone(id string) *Zone {
	return &Zone{id, c}
}
