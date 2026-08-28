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

const (
	braveDefaultBaseURL = "https://api.search.brave.com"
	engineBrave         = "brave"
	braveMaxResults     = 20
	braveMaxPage        = 10
)

var (
	// ErrBraveEmptyToken is returned when the API token is empty.
	ErrBraveEmptyToken = errors.New("brave: API token is required")
	// ErrBraveAPIStatus is returned when the API returns a non-2xx status.
	ErrBraveAPIStatus    = errors.New("brave: unexpected API status")
	errBraveInvalidLimit = errors.New("brave: limit must be between 0 and 20")
	errBraveInvalidPage  = errors.New("brave: page must be between 0 and 10")
)

// BraveProvider implements Provider for the Brave Web Search API.
type BraveProvider struct {
	token   string
	baseURL string
	client  *http.Client
}

// NewBraveProvider creates a Brave Web Search provider.
func NewBraveProvider(token string, options ...HTTPOption) *BraveProvider {
	return &BraveProvider{token: token, baseURL: braveDefaultBaseURL, client: newHTTPClient(options...)}
}

// CloseIdleConnections closes pooled provider connections.
func (p *BraveProvider) CloseIdleConnections() { p.client.CloseIdleConnections() }

// SetBaseURL overrides the API base URL for testing.
func (p *BraveProvider) SetBaseURL(value string) {
	p.baseURL = value
}

// Name returns the provider identifier.
func (p *BraveProvider) Name() string {
	return "brave"
}

// Engines returns the search engines supported by this provider.
func (p *BraveProvider) Engines() []string {
	return []string{engineBrave}
}

type braveResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Profile     struct {
		LongName string `json:"long_name"`
	} `json:"profile"`
}

type braveResponse struct {
	Web struct {
		Results []braveResult `json:"results"`
	} `json:"web"`
}

// Search performs a Brave Web Search request.
func (p *BraveProvider) Search(ctx context.Context, query string, opts Options) (*Response, error) {
	if p.token == "" {
		return nil, ErrBraveEmptyToken
	}
	if opts.Limit < 0 || opts.Limit > braveMaxResults {
		return nil, errBraveInvalidLimit
	}
	if opts.Page < 0 || opts.Page > braveMaxPage {
		return nil, errBraveInvalidPage
	}

	endpoint, err := p.buildURL(query, opts)
	if err != nil {
		return nil, err
	}
	headers := make(http.Header)
	headers.Set("X-Subscription-Token", p.token)
	body, err := httpGet(ctx, p.client, endpoint, "brave", ErrBraveAPIStatus, p.token, headers)
	if err != nil {
		return nil, err
	}

	var apiResponse braveResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("brave: parse response: %w", err)
	}
	results := make([]Result, len(apiResponse.Web.Results))
	for i, result := range apiResponse.Web.Results {
		results[i] = Result{
			Position: i + 1, Title: result.Title, URL: result.URL,
			Snippet: result.Description, DisplayedURL: result.Profile.LongName,
		}
	}
	response := &Response{Query: query, Engine: engineBrave, Provider: p.Name(), Results: results}
	if opts.Raw {
		if err := json.Unmarshal(body, &response.Raw); err != nil {
			return nil, fmt.Errorf("brave: parse raw response: %w", err)
		}
	}
	return response, nil
}

func (p *BraveProvider) buildURL(query string, opts Options) (string, error) {
	u, err := url.Parse(p.baseURL + "/res/v1/web/search")
	if err != nil {
		return "", fmt.Errorf("brave: parse base URL: %w", err)
	}
	values := u.Query()
	values.Set("q", query)
	if opts.Country != "" {
		values.Set("country", opts.Country)
	}
	if opts.Lang != "" {
		values.Set("search_lang", opts.Lang)
	}
	if opts.Limit > 0 {
		values.Set("count", strconv.Itoa(opts.Limit))
	}
	if opts.Page > 1 {
		values.Set("offset", strconv.Itoa(opts.Page-1))
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}
