---
name: scrape-do
description: |
  scrapedoctl provides web scraping, multi-engine search, URL mapping, and site crawling via MCP.

  USE SCRAPE-DO FOR:
  - Web scraping any URL to markdown (JS rendering, anti-bot bypass)
  - Multi-engine web search (Google, Bing, Yandex, DuckDuckGo, Baidu via SerpAPI)
  - URL discovery / site mapping
  - Recursive site crawling
  - Structured data extraction

  MCP TOOLS:
  - scrape_url: Scrape any URL to optimized markdown
  - web_search: Multi-engine search with provider selection
  - map_urls: Discover same-domain URLs on a page
  - crawl_site: Recursively crawl and extract site content

  CLI COMMANDS (when MCP tools are unavailable):
  - scrapedoctl search "query" --engine yandex
  - scrapedoctl scrape "url" --render
  - scrapedoctl map "url" --search "keyword"
  - scrapedoctl crawl "url" --depth 2 --limit 20
---

# Scrape.do CLI & MCP Server

Use the `scrape-do` MCP tools for all web scraping, search, mapping, and crawling tasks.

## MCP Tools

### scrape_url
Scrape any URL and return optimized markdown.
```
scrape_url(url: "https://example.com", render: true)
```
Parameters: `url` (required), `render`, `super`, `geoCode`, `device`, `headers`, `body`, `actions`.

### web_search
Search the web using multiple engines and providers.
```
web_search(query: "golang mcp sdk", engine: "google", limit: 10)
```
Parameters: `query` (required), `engine` (google/bing/yandex/duckduckgo/baidu/yahoo/naver), `provider` (scrapedo/serpapi/scraperapi), `lang`, `country`, `limit`.

### map_urls
Discover all same-domain URLs on a page.
```
map_urls(url: "https://docs.example.com", search: "api", limit: 50)
```
Parameters: `url` (required), `search` (filter keyword), `limit`.

### crawl_site
Recursively crawl a site and return all content.
```
crawl_site(url: "https://docs.example.com", maxDepth: 2, maxPages: 20)
```
Parameters: `url` (required), `maxDepth`, `maxPages`.

## CLI Fallback

If MCP tools are not available, use the CLI directly:

```bash
# Search
scrapedoctl search "query" --engine google --limit 5

# Scrape
scrapedoctl scrape "https://example.com" --render

# Map URLs
scrapedoctl map "https://example.com" --search "docs" --limit 20

# Crawl
scrapedoctl crawl "https://example.com" --depth 2 --limit 10 --output ./output

# Check provider usage
scrapedoctl account
scrapedoctl usage
```

## Provider Configuration

Configure search providers in `~/.scrapedoctl/conf.toml`:

```toml
[global]
token = "your-scrapedo-token"

[providers.serpapi]
token = "your-serpapi-key"
engines = ["google", "bing", "yandex", "duckduckgo", "baidu"]

[providers.scraperapi]
token = "your-scraperapi-key"
engines = ["google"]
```
