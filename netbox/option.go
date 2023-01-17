package netbox

import (
	"net/url"
)

// Option handles variadic configuration parameters passed to the
// client initializer.
type Option func(*Client) error

// WithNetBoxURL sets the URL, including the protocol and port (if
// specified) for the netbox installation.
func WithNetBoxURL(netboxURL string) Option {
	return func(c *Client) error {
		u, err := url.Parse(netboxURL)
		if err != nil {
			return err
		}
		c.baseURL = u
		return nil
	}
}

// WithToken sets the token used for requests.  This should generally
// be a read-only token as there is no reason this needs any kind of
// elevated permissions.
func WithToken(token string) Option {
	return func(c *Client) error {
		c.token = token
		return nil
	}
}
