package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const serpAPIDefaultBaseURL = "https://serpapi.com"

// Pagination parameter key constants.
const (
	serpPaginationStart = "start"
	serpPaginationFirst = "first"
	serpPaginationP     = "p"
	serpPaginationPN    = "pn"
	serpPaginationB     = "b"
)

// Sentinel errors for the SerpAPI provider.
var (
	// ErrSerpAPIEmptyToken is returned when the API token is empty.
	ErrSerpAPIEmptyToken = errors.New("serpapi: API token is required")
	// ErrSerpAPIStatus is returned when the API returns a non-200 status.
	ErrSerpAPIStatus = errors.New("serpapi: unexpected API status")
)

// serpAPIEngines lists all supported SerpAPI engine names.
var serpAPIEngines = []string{
	"google",
	"bing",
	"yandex",
	"duckduckgo",
	"baidu",
	"yahoo",
	"naver",
}

// SerpAPIProvider implements Provider for the SerpAPI multi-engine search service.
type SerpAPIProvider struct {
	token   string
	baseURL string
	client  *http.Client
	engines []string
}

// NewSerpAPIProvider creates a new SerpAPI search provider with the given API token.
func NewSerpAPIProvider(token string) *SerpAPIProvider {
	return &SerpAPIProvider{
		token:   token,
		baseURL: serpAPIDefaultBaseURL,
		client:  http.DefaultClient,
		engines: serpAPIEngines,
	}
}

// SetBaseURL overrides the API base URL (useful for testing).
func (p *SerpAPIProvider) SetBaseURL(u string) {
	p.baseURL = u
}

// Name returns the provider identifier.
func (p *SerpAPIProvider) Name() string {
	return "serpapi"
}

// Engines returns the search engines supported by this provider.
func (p *SerpAPIProvider) Engines() []string {
	return p.engines
}

// serpAPIOrganicResult maps a single item in the organic_results array.
type serpAPIOrganicResult struct {
	Position      int    `json:"position"`
	Title         string `json:"title"`
	Link          string `json:"link"`
	Snippet       string `json:"snippet"`
	DisplayedLink string `json:"displayed_link"`
}

// serpAPIResponse is the top-level JSON structure returned by SerpAPI.
type serpAPIResponse struct {
	SearchInformation map[string]any         `json:"search_information"`
	OrganicResults    []serpAPIOrganicResult `json:"organic_results"`
}

// Search performs a search via SerpAPI for the configured engine.
func (p *SerpAPIProvider) Search(ctx context.Context, query string, opts Options) (*Response, error) {
	if p.token == "" {
		return nil, ErrSerpAPIEmptyToken
	}

	engine := opts.Engine
	if engine == "" {
		engine = "google"
	}

	body, err := p.doRequest(ctx, query, engine, opts)
	if err != nil {
		return nil, err
	}

	return p.parseResponse(body, query, engine, opts)
}

func (p *SerpAPIProvider) doRequest(ctx context.Context, query, engine string, opts Options) ([]byte, error) {
	endpoint, err := p.buildURL(query, engine, opts)
	if err != nil {
		return nil, fmt.Errorf("serpapi: build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("serpapi: create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("serpapi: HTTP request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("serpapi: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w %d: %s", ErrSerpAPIStatus, resp.StatusCode, body)
	}

	return body, nil
}

func (p *SerpAPIProvider) parseResponse(body []byte, query, engine string, opts Options) (*Response, error) {
	var apiResp serpAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("serpapi: parse response: %w", err)
	}

	results := make([]Result, len(apiResp.OrganicResults))
	for i, r := range apiResp.OrganicResults {
		results[i] = Result{
			Position:     r.Position,
			Title:        r.Title,
			URL:          r.Link,
			Snippet:      r.Snippet,
			DisplayedURL: r.DisplayedLink,
		}
	}

	out := &Response{
		Query:    query,
		Engine:   engine,
		Provider: p.Name(),
		Results:  results,
		Metadata: map[string]any{},
	}

	if apiResp.SearchInformation != nil {
		if v, ok := apiResp.SearchInformation["total_results"]; ok {
			out.Metadata["total_results"] = v
		}
	}

	if opts.Raw {
		var raw any
		if err := json.Unmarshal(body, &raw); err != nil {
			return nil, fmt.Errorf("serpapi: parse raw response: %w", err)
		}
		out.Raw = raw
	}

	return out, nil
}

func (p *SerpAPIProvider) buildURL(query, engine string, opts Options) (string, error) {
	u, err := url.Parse(p.baseURL + "/search")
	if err != nil {
		return "", fmt.Errorf("serpapi: parse base URL: %w", err)
	}

	q := u.Query()
	q.Set("api_key", p.token)
	q.Set("engine", engine)
	q.Set(queryParam(engine), query)

	if opts.Lang != "" {
		q.Set("hl", opts.Lang)
	}
	if opts.Country != "" {
		q.Set("gl", opts.Country)
	}
	if opts.Page > 1 {
		key, val := paginationParam(engine, opts.Page)
		q.Set(key, val)
	}
	if opts.Limit > 0 {
		q.Set("num", strconv.Itoa(opts.Limit))
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// queryParam returns the query parameter name for the given engine.
func queryParam(engine string) string {
	switch engine {
	case "yandex":
		return "text"
	case "yahoo":
		return serpPaginationP
	case "naver":
		return "query"
	default:
		// google, bing, duckduckgo, baidu all use "q"
		return "q"
	}
}

// paginationParam returns the pagination parameter key and value for the given
// engine and 1-based page number.
func paginationParam(engine string, page int) (string, string) {
	offset := page - 1 // convert to 0-based

	switch engine {
	case "google":
		// start=0,10,20,...
		return serpPaginationStart, strconv.Itoa(offset * 10)
	case "bing":
		// first=1,11,21,...
		return serpPaginationFirst, strconv.Itoa(offset*10 + 1)
	case "yandex":
		// p=0,1,2,...
		return serpPaginationP, strconv.Itoa(offset)
	case "duckduckgo":
		// start=0,30,60,...
		return serpPaginationStart, strconv.Itoa(offset * 30)
	case "baidu":
		// pn=0,10,20,...
		return serpPaginationPN, strconv.Itoa(offset * 10)
	case "yahoo":
		// b=1,11,21,...
		return serpPaginationB, strconv.Itoa(offset*10 + 1)
	case "naver":
		// start=1,16,31,...
		return serpPaginationStart, strconv.Itoa(offset*15 + 1)
	default:
		return serpPaginationStart, strconv.Itoa(offset * 10)
	}
}
