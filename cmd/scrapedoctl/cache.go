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
				return fmt.Errorf("cache is disabled or not initialized")
			}

			stats, err := cacheStore.GetStats(context.Background())
			if err != nil {
				return err
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
				return fmt.Errorf("cache is disabled or not initialized")
			}

			if err := cacheStore.Clear(context.Background()); err != nil {
				return err
			}

			fmt.Println("Cache cleared successfully")
			return nil
		},
	}
}
