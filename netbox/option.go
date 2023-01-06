package netbox

import (
	"net/url"
)

type Option func(*Client) error

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

func WithToken(token string) Option {
	return func(c *Client) error {
		c.token = token
		return nil
	}
}
