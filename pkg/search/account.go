package search

import "context"

// AccountChecker is an optional interface providers can implement to report usage.
type AccountChecker interface {
	Account(ctx context.Context) (*AccountInfo, error)
}

// AccountInfo holds normalized account usage information for a provider.
type AccountInfo struct {
	Provider          string `json:"provider"`
	Plan              string `json:"plan,omitempty"`
	Active            bool   `json:"active"`
	UsedRequests      int    `json:"used_requests"`
	MaxRequests       int    `json:"max_requests"`
	RemainingRequests int    `json:"remaining_requests"`
	Concurrency       int    `json:"concurrency,omitempty"`
	RateLimit         int    `json:"rate_limit_per_hour,omitempty"`
}
