package search

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

// Sentinel errors returned by the Router.
var (
	// ErrProviderNotSupported is returned when a named provider does not support the requested engine.
	ErrProviderNotSupported = errors.New("provider does not support engine")
	// ErrNoProvider is returned when no registered provider supports the requested engine.
	ErrNoProvider = errors.New("no provider found for engine")
)

// Router manages provider registration and resolves which provider handles a given engine.
type Router struct {
	providers []Provider
}

// NewRouter creates an empty Router.
func NewRouter() *Router {
	return &Router{}
}

// Register adds a provider to the router.
func (r *Router) Register(p Provider) {
	r.providers = append(r.providers, p)
}

// Providers returns all registered providers.
func (r *Router) Providers() []Provider {
	return r.providers
}

// Resolve finds a provider for the given engine. If providerName is non-empty,
// only that provider is considered.
func (r *Router) Resolve(engine, providerName string) (Provider, error) {
	for _, p := range r.providers {
		if providerName != "" && p.Name() != providerName {
			continue
		}
		if slices.Contains(p.Engines(), engine) {
			return p, nil
		}
	}
	if providerName != "" {
		return nil, fmt.Errorf("%w: provider %q engine %q", ErrProviderNotSupported, providerName, engine)
	}
	return nil, fmt.Errorf("%w %q (available: %v)", ErrNoProvider, engine, r.AllEngines())
}

// Search resolves a provider and performs the search.
func (r *Router) Search(ctx context.Context, query string, opts Options) (*Response, error) {
	p, err := r.Resolve(opts.Engine, "")
	if err != nil {
		return nil, err
	}
	resp, err := p.Search(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("provider %q search: %w", p.Name(), err)
	}
	return resp, nil
}

// AllEngines returns a deduplicated list of all engines across all providers.
func (r *Router) AllEngines() []string {
	seen := make(map[string]bool)
	var engines []string
	for _, p := range r.providers {
		for _, e := range p.Engines() {
			if !seen[e] {
				seen[e] = true
				engines = append(engines, e)
			}
		}
	}
	return engines
}

// ProviderNames returns names of all registered providers.
func (r *Router) ProviderNames() []string {
	names := make([]string, len(r.providers))
	for i, p := range r.providers {
		names[i] = p.Name()
	}
	return names
}
