package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/ioplane/scrapedoctl/internal/config"
)

// knownProviders maps provider names to their default engines.
var knownProviders = map[string][]string{
	"serpapi":    {"google", "bing", "yandex", "duckduckgo", "baidu", "yahoo", "naver"},
	"scraperapi": {"google"},
	"brave":      {"brave"},
}

// errUnknownProvider is returned when a provider name is not recognized.
var errUnknownProvider = fmt.Errorf("unknown provider (known: %s)", knownProviderNames())

// errProviderNotConfigured is returned when trying to remove a provider that is not configured.
var errProviderNotConfigured = errors.New("provider not configured")

func knownProviderNames() string {
	names := make([]string, 0, len(knownProviders))
	for n := range knownProviders {
		names = append(names, n)
	}
	sort.Strings(names)

	return strings.Join(names, ", ")
}

func newProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage search providers",
	}

	cmd.AddCommand(newProviderListCmd())
	cmd.AddCommand(newProviderAddCmd())
	cmd.AddCommand(newProviderRemoveCmd())

	return cmd
}

func newProviderListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List configured search providers",
		RunE: func(_ *cobra.Command, _ []string) error {
			if jsonOutput {
				return printProvidersJSON(cfg)
			}

			printProvidersTable(cfg)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newProviderAddCmd() *cobra.Command {
	var tokenEnvironment string
	var tokenStdin bool

	cmd := &cobra.Command{
		Use:   "add <provider>",
		Short: "Add a search provider",
		Example: "  scrapedoctl provider add serpapi --token-env SERPAPI_TOKEN\n" +
			"  scrapedoctl provider add brave --token-stdin",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, ok := knownProviders[args[0]]; !ok {
				return fmt.Errorf("%w: %s", errUnknownProvider, args[0])
			}
			token, err := readSecret(cmd, tokenEnvironment, tokenStdin, args[0]+" API Token")
			if err != nil {
				return err
			}
			return runProviderAdd(args[0], token)
		},
	}

	cmd.Flags().StringVar(&tokenEnvironment, "token-env", "", "Read the token from this environment variable")
	cmd.Flags().BoolVar(&tokenStdin, "token-stdin", false, "Read the token from stdin")
	cmd.MarkFlagsMutuallyExclusive("token-env", "token-stdin")

	return cmd
}

func runProviderAdd(name, token string) error {
	engines, ok := knownProviders[name]
	if !ok {
		return fmt.Errorf("%w: %s", errUnknownProvider, name)
	}

	if cfg.Providers == nil {
		cfg.Providers = make(map[string]config.ProviderConfig)
	}

	cfg.Providers[name] = config.ProviderConfig{
		Token:   token,
		Engines: engines,
	}

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Added provider %q with engines: %s\n", name, strings.Join(engines, ", "))

	return nil
}

func newProviderRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <provider>",
		Short: "Remove a search provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runProviderRemove(args[0])
		},
	}
}

func runProviderRemove(name string) error {
	if cfg.Providers == nil {
		return fmt.Errorf("%w: %s", errProviderNotConfigured, name)
	}

	if _, ok := cfg.Providers[name]; !ok {
		return fmt.Errorf("%w: %s", errProviderNotConfigured, name)
	}

	delete(cfg.Providers, name)

	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Removed provider %q\n", name)

	return nil
}

// providerRow holds display data for one provider.
type providerRow struct {
	Name    string   `json:"name"`
	Engines []string `json:"engines"`
	Status  string   `json:"status"`
}

func buildProviderRows(c *config.Config) []providerRow {
	var rows []providerRow

	// Always show scrapedo first if token is set.
	if c.Global.Token != "" {
		rows = append(rows, providerRow{
			Name:    "scrapedo",
			Engines: []string{"google"},
			Status:  "active (global token)",
		})
	}

	names := make([]string, 0, len(c.Providers))
	for name := range c.Providers {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		pcfg := c.Providers[name]
		if pcfg.Type != "exec" {
			if _, supported := knownProviders[name]; !supported {
				continue
			}
		}
		status := "active"
		if pcfg.Token == "" && pcfg.Command == "" {
			status = "no token"
		}

		rows = append(rows, providerRow{
			Name:    name,
			Engines: pcfg.Engines,
			Status:  status,
		})
	}

	return rows
}

func printProvidersTable(c *config.Config) {
	rows := buildProviderRows(c)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Provider\tEngines\tStatus")

	for _, r := range rows {
		fmt.Fprintf(w, "%s\t%s\t%s\n", r.Name, strings.Join(r.Engines, ", "), r.Status)
	}

	if err := w.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "flush: %v\n", err)
	}
}

func printProvidersJSON(c *config.Config) error {
	rows := buildProviderRows(c)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")

	if err := enc.Encode(rows); err != nil {
		return fmt.Errorf("encode providers: %w", err)
	}

	return nil
}
