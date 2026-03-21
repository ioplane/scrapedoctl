// Package main provides the entry point for scrapedoctl.
package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage the persistent cache",
	}

	cmd.AddCommand(newCacheStatsCmd())
	cmd.AddCommand(newCacheClearCmd())

	return cmd
}

func newCacheStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show cache statistics",
		RunE: func(_ *cobra.Command, _ []string) error {
			if cacheStore == nil {
				return errCacheNotInitialized
			}

			stats, err := cacheStore.GetStats(context.Background())
			if err != nil {
				return fmt.Errorf("failed to get stats: %w", err)
			}

			fmt.Printf("Total entries: %d\n", stats.TotalCount)
			fmt.Printf("Total size:    %.2f MB\n", float64(stats.TotalSize)/(1024*1024))
			return nil
		},
	}
}

func newCacheClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear all cached results",
		RunE: func(_ *cobra.Command, _ []string) error {
			if cacheStore == nil {
				return errCacheNotInitialized
			}

			if err := cacheStore.Clear(context.Background()); err != nil {
				return fmt.Errorf("failed to clear cache: %w", err)
			}

			fmt.Println("Cache cleared successfully")
			return nil
		},
	}
}
