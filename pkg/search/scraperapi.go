package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const scraperAPIDefaultBaseURL = "https://api.scraperapi.com"

// Sentinel errors for ScraperAPIProvider.
var (
	// ErrScraperAPIEmptyToken is returned when the API token is empty.
	ErrScraperAPIEmptyToken = errors.New("scraperapi: API token is required")
	// ErrScraperAPIStatus is returned when the API returns a non-200 status.
	ErrScraperAPIStatus = errors.New("scraperapi: unexpected API status")
)

// ScraperAPIProvider implements Provider for the ScraperAPI Google Search API.
type ScraperAPIProvider struct {
	token   string
	baseURL string
	client  *http.Client
}

// NewScraperAPIProvider creates a new ScraperAPI search provider with the given API token.
func NewScraperAPIProvider(token string) *ScraperAPIProvider {
	return &ScraperAPIProvider{
		token:   token,
		baseURL: scraperAPIDefaultBaseURL,
		client:  http.DefaultClient,
	}
}

// SetBaseURL overrides the API base URL (useful for testing).
func (p *ScraperAPIProvider) SetBaseURL(u string) {
	p.baseURL = u
}

// Name returns the provider identifier.
func (p *ScraperAPIProvider) Name() string {
	return "scraperapi"
}

// Engines returns the search engines supported by this provider.
func (p *ScraperAPIProvider) Engines() []string {
	return []string{"google"}
}

// scraperAPISearchInfo maps the search_information block in the API response.
type scraperAPISearchInfo struct {
	QueryDisplayed string `json:"query_displayed"`
}

// scraperAPIOrganicResult maps a single item in the organic_results array.
type scraperAPIOrganicResult struct {
	Position      int    `json:"position"`
	Title         string `json:"title"`
	Link          string `json:"link"`
	Snippet       string `json:"snippet"`
	DisplayedLink string `json:"displayed_link"`
}

// scraperAPIPagination maps the pagination block in the API response.
type scraperAPIPagination struct {
	PagesCount  int    `json:"pages_count"`
	CurrentPage int    `json:"current_page"`
	NextPageURL string `json:"next_page_url"`
}

// scraperAPIResponse is the top-level JSON structure returned by the API.
type scraperAPIResponse struct {
	SearchInformation scraperAPISearchInfo      `json:"search_information"`
	OrganicResults    []scraperAPIOrganicResult `json:"organic_results"`
	Pagination        scraperAPIPagination      `json:"pagination"`
}

// Search performs a Google search via the ScraperAPI structured endpoint.
func (p *ScraperAPIProvider) Search(ctx context.Context, query string, opts Options) (*Response, error) {
	if p.token == "" {
		return nil, ErrScraperAPIEmptyToken
	}

	body, err := p.doRequest(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	var apiResp scraperAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("scraperapi: parse response: %w", err)
	}

	return p.buildResponse(query, opts, body, &apiResp), nil
}

func (p *ScraperAPIProvider) doRequest(ctx context.Context, query string, opts Options) ([]byte, error) {
	endpoint, err := p.buildURL(query, opts)
	if err != nil {
		return nil, fmt.Errorf("scraperapi: build URL: %w", err)
	}

	return httpGet(ctx, p.client, endpoint, "scraperapi", ErrScraperAPIStatus)
}

func (p *ScraperAPIProvider) buildResponse(
	query string, opts Options, body []byte, apiResp *scraperAPIResponse,
) *Response {
	results := make([]Result, len(apiResp.OrganicResults))
	for i, r := range apiResp.OrganicResults {
		results[i] = Result{
			Position:     r.Position + 1,
			Title:        r.Title,
			URL:          r.Link,
			Snippet:      r.Snippet,
			DisplayedURL: r.DisplayedLink,
		}
	}

	out := &Response{
		Query:    query,
		Engine:   "google",
		Provider: p.Name(),
		Results:  results,
		Metadata: map[string]any{
			"query_displayed": apiResp.SearchInformation.QueryDisplayed,
			"pages_count":     apiResp.Pagination.PagesCount,
			"current_page":    apiResp.Pagination.CurrentPage,
		},
	}

	if opts.Raw {
		var raw any
		if err := json.Unmarshal(body, &raw); err == nil {
			out.Raw = raw
		}
	}

	return out
}

// scraperAPIAccountResponse maps the JSON from the /account endpoint.
type scraperAPIAccountResponse struct {
	ConcurrencyLimit int `json:"concurrencyLimit"`
	RequestCount     int `json:"requestCount"`
	RequestLimit     int `json:"requestLimit"`
}

// Account retrieves account usage information from the ScraperAPI.
func (p *ScraperAPIProvider) Account(ctx context.Context) (*AccountInfo, error) {
	if p.token == "" {
		return nil, ErrScraperAPIEmptyToken
	}

	endpoint := p.baseURL + "/account?api_key=" + url.QueryEscape(p.token)

	body, err := httpGet(ctx, p.client, endpoint, "scraperapi", ErrScraperAPIStatus)
	if err != nil {
		return nil, err
	}

	var resp scraperAPIAccountResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("scraperapi: parse account response: %w", err)
	}

	remaining := resp.RequestLimit - resp.RequestCount

	return &AccountInfo{
		Provider:          p.Name(),
		Active:            true,
		UsedRequests:      resp.RequestCount,
		MaxRequests:       resp.RequestLimit,
		RemainingRequests: remaining,
		Concurrency:       resp.ConcurrencyLimit,
	}, nil
}

func (p *ScraperAPIProvider) buildURL(query string, opts Options) (string, error) {
	u, err := url.Parse(p.baseURL + "/structured/google/search")
	if err != nil {
		return "", fmt.Errorf("scraperapi: parse base URL: %w", err)
	}

	q := u.Query()
	q.Set("api_key", p.token)
	q.Set("query", query)

	if opts.Country != "" {
		q.Set("country_code", opts.Country)
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}
