package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const scrapedoDefaultBaseURL = "https://api.scrape.do"

// Sentinel errors for ScrapedoProvider.
var (
	// ErrScrapedoEmptyToken is returned when the API token is empty.
	ErrScrapedoEmptyToken = errors.New("scrapedo: API token is required")
	// ErrScrapedoAPIStatus is returned when the API returns a non-200 status.
	ErrScrapedoAPIStatus = errors.New("scrapedo: unexpected API status")
)

// ScrapedoProvider implements Provider for the Scrape.do Google Search API.
type ScrapedoProvider struct {
	token   string
	baseURL string
	client  *http.Client
}

// NewScrapedoProvider creates a new Scrape.do search provider with the given API token.
func NewScrapedoProvider(token string) *ScrapedoProvider {
	return &ScrapedoProvider{
		token:   token,
		baseURL: scrapedoDefaultBaseURL,
		client:  http.DefaultClient,
	}
}

// SetBaseURL overrides the API base URL (useful for testing).
func (p *ScrapedoProvider) SetBaseURL(u string) {
	p.baseURL = u
}

// Name returns the provider identifier.
func (p *ScrapedoProvider) Name() string {
	return "scrapedo"
}

// Engines returns the search engines supported by this provider.
func (p *ScrapedoProvider) Engines() []string {
	return []string{"google"}
}

// scrapedoSearchInfo maps the search_information block in the API response.
type scrapedoSearchInfo struct {
	TotalResults       any `json:"total_results"`
	TimeTakenDisplayed any `json:"time_taken_displayed"`
}

// scrapedoOrganicResult maps a single item in the organic_results array.
type scrapedoOrganicResult struct {
	Position      int    `json:"position"`
	Title         string `json:"title"`
	Link          string `json:"link"`
	Snippet       string `json:"snippet"`
	DisplayedLink string `json:"displayed_link"`
}

// scrapedoResponse is the top-level JSON structure returned by the API.
type scrapedoResponse struct {
	SearchInformation scrapedoSearchInfo      `json:"search_information"`
	OrganicResults    []scrapedoOrganicResult `json:"organic_results"`
}

// Search performs a Google search via the Scrape.do API.
func (p *ScrapedoProvider) Search(ctx context.Context, query string, opts Options) (*Response, error) {
	if p.token == "" {
		return nil, ErrScrapedoEmptyToken
	}

	body, err := p.doRequest(ctx, query, opts)
	if err != nil {
		return nil, err
	}

	var apiResp scrapedoResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("scrapedo: parse response: %w", err)
	}

	return p.buildResponse(query, opts, body, &apiResp), nil
}

func (p *ScrapedoProvider) doRequest(ctx context.Context, query string, opts Options) ([]byte, error) {
	endpoint, err := p.buildURL(query, opts)
	if err != nil {
		return nil, fmt.Errorf("scrapedo: build URL: %w", err)
	}

	return httpGet(ctx, p.client, endpoint, "scrapedo", ErrScrapedoAPIStatus)
}

func (p *ScrapedoProvider) buildResponse(
	query string, opts Options, body []byte, apiResp *scrapedoResponse,
) *Response {
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
		Engine:   "google",
		Provider: p.Name(),
		Results:  results,
		Metadata: map[string]any{
			"total_results": apiResp.SearchInformation.TotalResults,
			"time_taken":    apiResp.SearchInformation.TimeTakenDisplayed,
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

// scrapedoAccountResponse maps the JSON from the /info endpoint.
type scrapedoAccountResponse struct {
	IsActive                   bool `json:"IsActive"`
	ConcurrentRequest          int  `json:"ConcurrentRequest"`
	MaxMonthlyRequest          int  `json:"MaxMonthlyRequest"`
	RemainingConcurrentRequest int  `json:"RemainingConcurrentRequest"`
	RemainingMonthlyRequest    int  `json:"RemainingMonthlyRequest"`
}

// Account retrieves account usage information from the Scrape.do API.
func (p *ScrapedoProvider) Account(ctx context.Context) (*AccountInfo, error) {
	if p.token == "" {
		return nil, ErrScrapedoEmptyToken
	}

	endpoint := p.baseURL + "/info?token=" + url.QueryEscape(p.token)

	body, err := httpGet(ctx, p.client, endpoint, "scrapedo", ErrScrapedoAPIStatus)
	if err != nil {
		return nil, err
	}

	var resp scrapedoAccountResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("scrapedo: parse account response: %w", err)
	}

	used := resp.MaxMonthlyRequest - resp.RemainingMonthlyRequest

	return &AccountInfo{
		Provider:          p.Name(),
		Active:            resp.IsActive,
		UsedRequests:      used,
		MaxRequests:       resp.MaxMonthlyRequest,
		RemainingRequests: resp.RemainingMonthlyRequest,
		Concurrency:       resp.ConcurrentRequest,
	}, nil
}

func (p *ScrapedoProvider) buildURL(query string, opts Options) (string, error) {
	u, err := url.Parse(p.baseURL + "/plugin/google/search")
	if err != nil {
		return "", fmt.Errorf("scrapedo: parse base URL: %w", err)
	}

	q := u.Query()
	q.Set("token", p.token)
	q.Set("q", query)

	if opts.Lang != "" {
		q.Set("hl", opts.Lang)
	}
	if opts.Country != "" {
		q.Set("gl", opts.Country)
	}
	if opts.Page > 1 {
		q.Set("start", strconv.Itoa((opts.Page-1)*10))
	}
	if opts.Limit > 0 {
		q.Set("num", strconv.Itoa(opts.Limit))
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}
