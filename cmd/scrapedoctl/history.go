package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newHistoryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history <url>",
		Short: "Show scrape history for a URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cacheStore == nil {
				return fmt.Errorf("cache is disabled or not initialized")
			}

			url := args[0]
			history, err := cacheStore.GetHistory(context.Background(), url)
			if err != nil {
				return fmt.Errorf("failed to fetch history: %w", err)
			}

			if len(history) == 0 {
				fmt.Printf("No history found for %s\n", url)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tDATE\tCOST\tCREDITS\tSTATUS")
			for _, r := range history {
				var meta map[string]any
				_ = json.Unmarshal([]byte(r.Metadata), &meta)
				fmt.Fprintf(w, "%d\t%s\t%v\t%v\t%v\n",
					r.ID,
					r.CreatedAt.Format("2006-01-02 15:04:05"),
					meta["cost"],
					meta["remaining_credits"],
					meta["status"],
				)
			}
			w.Flush()
			return nil
		},
	}
}
