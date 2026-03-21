/*
Package scrapedo provides an HTTP client for the Scrape.do web scraping API.

# Client

Create a client with [NewClient] and use [Client.Scrape] to fetch pages:

	client, err := scrapedo.NewClient("your-api-token")
	if err != nil {
		log.Fatal(err)
	}

	content, err := client.Scrape(ctx, scrapedo.ScrapeRequest{
		URL:    "https://example.com",
		Render: true,
	})

# Features

The client supports:
  - JavaScript rendering (headless browser)
  - Residential proxy rotation
  - Geo-targeting by country code
  - Sticky sessions
  - Device emulation (desktop, mobile, tablet)
  - Custom headers and POST requests
  - Browser actions (clicks, scrolling)

# URL Discovery

Use [ExtractLinks] to discover same-domain URLs from scraped content,
and [Client.Crawl] for recursive site crawling with BFS.
*/
package scrapedo
