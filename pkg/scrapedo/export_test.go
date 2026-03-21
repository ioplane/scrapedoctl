package scrapedo

import "net/url"

// MaskTokenInURL is an internal method exported for testing.
func (c *Client) MaskTokenInURL(u *url.URL) string {
	return c.maskTokenInURL(u)
}

// PrepareQueryParams is an internal method exported for testing.
func (c *Client) PrepareQueryParams(req ScrapeRequest) (url.Values, error) {
	return c.prepareQueryParams(req)
}
