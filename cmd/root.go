// Package cmd contains all CLI commands for podfetch.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "podfetch",
	Short: "A CLI to fetch and filter podcast episodes by duration",
	Long: `podfetch is a command-line tool to track podcast feeds and find
episodes that fit your available time. Perfect for finding a 30-minute
episode for your commute or a 2-hour deep-dive for a long drive.

Add feeds, fetch episodes, and search by duration:
  podfetch add https://example.com/feed.xml
  podfetch fetch
  podfetch search --max-duration 45m`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
