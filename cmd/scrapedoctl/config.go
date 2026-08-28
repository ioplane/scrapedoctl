// Package main provides the entry point for scrapedoctl.
package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ioplane/scrapedoctl/internal/config"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage scrapedoctl settings",
	}

	cmd.AddCommand(newConfigListCmd())
	cmd.AddCommand(newConfigSetCmd())

	return cmd
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all current settings",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Global Token: %s\n", config.RedactedSecret(cfg.Global.Token))
			fmt.Printf("Global BaseURL: %s\n", cfg.Global.BaseURL)
			fmt.Printf("Global Timeout: %d\n", cfg.Global.Timeout)
			fmt.Printf("REPL History: %s\n", cfg.Repl.HistoryFile)
			fmt.Printf("Active Profile: %s\n", cfg.ActiveProfile)
			fmt.Println("\nResolved Settings (including profile/env):")
			fmt.Printf("  Render: %v\n", cfg.Resolved.Render)
			fmt.Printf("  Super: %v\n", cfg.Resolved.Super)
			fmt.Printf("  GeoCode: %s\n", cfg.Resolved.GeoCode)
			fmt.Printf("  Device: %s\n", cfg.Resolved.Device)
			fmt.Printf("  Session: %s\n", cfg.Resolved.Session)
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	var fromEnvironment string
	var fromStdin bool

	cmd := &cobra.Command{
		Use:   "set <key>[=<value>]",
		Short: "Set a configuration value",
		Example: "  scrapedoctl config set global.token --from-env SCRAPEDO_TOKEN\n" +
			"  scrapedoctl config set global.token --from-stdin",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key, value, err := configValue(cmd, args[0], fromEnvironment, fromStdin)
			if err != nil {
				return err
			}
			previous := *cfg

			switch key {
			case "global.token":
				cfg.Global.Token = value
			case "global.base_url":
				cfg.Global.BaseURL = value
			case "global.timeout":
				timeout, err := strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("%w: global.timeout must be an integer", config.ErrInvalidConfig)
				}
				cfg.Global.Timeout = timeout
			case "repl.history_file":
				cfg.Repl.HistoryFile = value
			default:
				return fmt.Errorf("%w: %s", errUnsupportedConfigKey, key)
			}
			if err := cfg.Validate(); err != nil {
				*cfg = previous
				return fmt.Errorf("validate config: %w", err)
			}

			if err := cfg.Save(); err != nil {
				*cfg = previous
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Successfully set %s\n", key)
			return nil
		},
	}

	cmd.Flags().StringVar(&fromEnvironment, "from-env", "", "Read global.token from this environment variable")
	cmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read global.token from stdin")
	cmd.MarkFlagsMutuallyExclusive("from-env", "from-stdin")
	return cmd
}

func configValue(cmd *cobra.Command, argument, environment string, stdin bool) (string, string, error) {
	if strings.HasPrefix(argument, "global.token=") {
		return "", "", errSecretInArguments
	}
	if argument == "global.token" {
		value, err := readSecret(cmd, environment, stdin, "Scrape.do API Token")
		return argument, value, err
	}
	if environment != "" || stdin {
		return "", "", errSecretSourceOnlyToken
	}
	parts := strings.SplitN(argument, "=", 2)
	if len(parts) != 2 {
		return "", "", errInvalidConfigFormat
	}
	return parts[0], parts[1], nil
}
