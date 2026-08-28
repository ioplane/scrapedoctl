package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/ioplane/scrapedoctl/internal/cache"
	"github.com/ioplane/scrapedoctl/internal/config"
	"github.com/ioplane/scrapedoctl/pkg/scrapedo"
)

const maxSecretBytes = int64(4 << 10)

// buildClient creates a scrapedo.Client from config/env and attaches the cache.
func buildClient(
	c *config.Config, store *cache.Store, extraOptions ...scrapedo.ClientOption,
) (*scrapedo.Client, error) {
	token := c.Global.Token
	if token == "" {
		token = os.Getenv("SCRAPEDO_TOKEN")
	}

	if token == "" {
		return nil, errMissingToken
	}

	options := make([]scrapedo.ClientOption, 0, 3+len(extraOptions))
	if c.Global.BaseURL != "" {
		options = append(options, scrapedo.WithBaseURL(c.Global.BaseURL))
	}
	if c.Global.Timeout > 0 {
		options = append(options, scrapedo.WithTimeout(time.Duration(c.Global.Timeout)*time.Millisecond))
	}
	if store != nil {
		options = append(options, scrapedo.WithCache(store))
	}
	options = append(options, extraOptions...)

	client, err := scrapedo.NewClient(token, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}

// commandContext supports direct RunE unit calls while production execution uses ExecuteContext.
func commandContext(cmd interface{ Context() context.Context }) context.Context {
	if ctx := cmd.Context(); ctx != nil {
		return ctx
	}
	return context.Background()
}

func readSecret(cmd *cobra.Command, environment string, stdin bool, title string) (string, error) {
	var secret string
	switch {
	case environment != "":
		var ok bool
		secret, ok = os.LookupEnv(environment)
		if !ok {
			return "", fmt.Errorf("%w: %q", errSecretEnvironmentUnset, environment)
		}
	case stdin:
		value, err := io.ReadAll(io.LimitReader(cmd.InOrStdin(), maxSecretBytes+1))
		if err != nil {
			return "", fmt.Errorf("read secret from stdin: %w", err)
		}
		if int64(len(value)) > maxSecretBytes {
			return "", errSecretTooLarge
		}
		secret = string(value)
	default:
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title(title).Value(&secret).EchoMode(huh.EchoModePassword),
		)).WithInput(cmd.InOrStdin()).WithOutput(cmd.OutOrStdout())
		if err := form.RunWithContext(commandContext(cmd)); err != nil {
			return "", fmt.Errorf("read secret from hidden prompt: %w", err)
		}
	}

	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", errSecretEmpty
	}
	return secret, nil
}

// extractHost returns the hostname from a URL string.
func extractHost(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	return u.Host
}

// sanitizePath converts a URL path to a safe filename.
var unsafePathChars = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func sanitizePath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "page"
	}

	p := strings.Trim(u.Path, "/")
	if p == "" {
		return "index"
	}

	return unsafePathChars.ReplaceAllString(p, "_")
}
