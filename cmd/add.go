package cmd

import (
	"errors"
	"fmt"

	"github.com/johnmaher/podfetch/internal/feed"
	"github.com/johnmaher/podfetch/internal/store"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <rss-url>",
	Short: "Add a podcast feed to track",
	Long:  `Add a podcast RSS feed URL to your tracked feeds list.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	url := args[0]

	fmt.Printf("Fetching feed from %s...\n", url)

	// Fetch the feed to validate and get the title
	f, err := feed.Fetch(url)
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}

	// Initialize store
	s, err := store.New()
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	// Add feed to tracked list
	if err := s.AddFeed(url, f.Title); err != nil {
		if errors.Is(err, store.ErrFeedExists) {
			return fmt.Errorf("feed already tracked: %s", f.Title)
		}
		return fmt.Errorf("failed to save feed: %w", err)
	}

	// Cache the feed data
	if err := s.UpdateCache(f); err != nil {
		return fmt.Errorf("failed to cache feed: %w", err)
	}

	fmt.Printf("✓ Added: %s (%d episodes)\n", f.Title, len(f.Episodes))
	return nil
}
