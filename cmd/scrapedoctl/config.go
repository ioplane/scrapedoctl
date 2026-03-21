package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
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
			fmt.Printf("Global Token: %s\n", cfg.Global.Token)
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
	return &cobra.Command{
		Use:   "set <key>=<value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			parts := strings.SplitN(args[0], "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid format, use key=value")
			}
			key, value := parts[0], parts[1]

			switch key {
			case "global.token":
				cfg.Global.Token = value
			case "global.base_url":
				cfg.Global.BaseURL = value
			case "repl.history_file":
				cfg.Repl.HistoryFile = value
			default:
				return fmt.Errorf("unknown or unsupported key: %s", key)
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Printf("Successfully set %s\n", key)
			return nil
		},
	}
}
