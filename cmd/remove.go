package cmd

import (
	"errors"
	"fmt"

	"github.com/johnmaher/podfetch/internal/store"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <feed-name-or-url>",
	Short: "Remove a tracked feed",
	Long:  `Remove a podcast feed from your tracked list by name or URL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	identifier := args[0]

	s, err := store.New()
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	if err := s.RemoveFeed(identifier); err != nil {
		if errors.Is(err, store.ErrFeedNotFound) {
			return fmt.Errorf("feed not found: %s", identifier)
		}
		return fmt.Errorf("failed to remove feed: %w", err)
	}

	fmt.Printf("✓ Removed: %s\n", identifier)
	return nil
}
