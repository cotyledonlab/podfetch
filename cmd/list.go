package cmd

import (
	"fmt"

	"github.com/johnmaher/podfetch/internal/store"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tracked podcast feeds",
	Long:  `Display all podcast feeds currently being tracked.`,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	data, err := s.LoadFeeds()
	if err != nil {
		return fmt.Errorf("failed to load feeds: %w", err)
	}

	if len(data.Feeds) == 0 {
		fmt.Println("No feeds tracked yet. Add one with: podfetch add <url>")
		return nil
	}

	fmt.Printf("Tracking %d feed(s):\n\n", len(data.Feeds))
	for i, f := range data.Feeds {
		fmt.Printf("%d. %s\n", i+1, f.Title)
		fmt.Printf("   %s\n\n", f.URL)
	}

	return nil
}
