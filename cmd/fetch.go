package cmd

import (
	"fmt"
	"sync/atomic"

	"github.com/johnmaher/podfetch/internal/feed"
	"github.com/johnmaher/podfetch/internal/store"
	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Refresh all tracked feeds",
	Long:  `Fetch the latest episodes from all tracked podcast feeds concurrently.`,
	RunE:  runFetch,
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}

func runFetch(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	data, err := s.LoadFeeds()
	if err != nil {
		return fmt.Errorf("failed to load feeds: %w", err)
	}

	if len(data.Feeds) == 0 {
		fmt.Println("No feeds to fetch. Add one with: podfetch add <url>")
		return nil
	}

	fmt.Printf("Fetching %d feed(s) concurrently...\n\n", len(data.Feeds))

	// Collect URLs
	urls := make([]string, len(data.Feeds))
	for i, f := range data.Feeds {
		urls[i] = f.URL
	}

	// Fetch all feeds concurrently using goroutines + channels
	results := feed.FetchConcurrently(urls)

	// Process results as they come in
	var successCount, errorCount int32
	for result := range results {
		if result.Error != nil {
			atomic.AddInt32(&errorCount, 1)
			fmt.Printf("✗ Error fetching feed: %v\n", result.Error)
			continue
		}

		// Update cache
		if err := s.UpdateCache(result.Feed); err != nil {
			atomic.AddInt32(&errorCount, 1)
			fmt.Printf("✗ Error caching %s: %v\n", result.Feed.Title, err)
			continue
		}

		atomic.AddInt32(&successCount, 1)
		fmt.Printf("✓ %s (%d episodes)\n", result.Feed.Title, len(result.Feed.Episodes))
	}

	fmt.Printf("\nFetch complete: %d succeeded, %d failed\n", successCount, errorCount)
	return nil
}
