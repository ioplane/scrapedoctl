/*
Package search provides a pluggable multi-provider web search system.

# Providers

The search system supports multiple providers through the [Provider] interface:
  - Scrape.do (Google) via [NewScrapedoProvider]
  - ScraperAPI (Google) via [NewScraperAPIProvider]
  - SerpAPI (7 engines) via [NewSerpAPIProvider]
  - Custom exec plugins via [NewExecProvider]

# Router

The [Router] dispatches search queries to the appropriate provider
based on engine support. It resolves providers by engine name and
supports explicit provider selection:

	router := search.NewRouter()
	router.Register(search.NewScrapedoProvider("scrapedo-token"))
	router.Register(search.NewSerpAPIProvider("serpapi-key"))

	resp, err := router.Search(ctx, "golang mcp sdk", search.Options{
		Engine: "yandex",
		Limit:  10,
	})

# Output Formats

Results can be formatted as table, JSON, or Markdown:
  - [FormatTable] — tabular output for terminals
  - [FormatJSON] — machine-readable JSON
  - [FormatMarkdown] — LLM-optimized markdown with links

# Account Information

Providers implementing [AccountChecker] report usage, limits, and
credits via [AccountInfo].
*/
package search
