package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/johnmaher/podfetch/internal/feed"
	"github.com/johnmaher/podfetch/internal/store"
	"github.com/spf13/cobra"
)

var (
	maxDuration string
	minDuration string
	query       string
	limit       int
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search episodes by duration and keyword",
	Long: `Search across all cached episodes with filters for duration and keywords.

Examples:
  podfetch search --max-duration 30m
  podfetch search --min-duration 1h --max-duration 2h
  podfetch search --query "AI" --max-duration 45m`,
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringVar(&maxDuration, "max-duration", "", "Maximum episode duration (e.g., 30m, 1h30m)")
	searchCmd.Flags().StringVar(&minDuration, "min-duration", "", "Minimum episode duration (e.g., 15m)")
	searchCmd.Flags().StringVarP(&query, "query", "q", "", "Search keyword in title/description")
	searchCmd.Flags().IntVarP(&limit, "limit", "n", 20, "Maximum number of results to show")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	s, err := store.New()
	if err != nil {
		return fmt.Errorf("failed to initialize store: %w", err)
	}

	episodes, err := s.GetAllEpisodes()
	if err != nil {
		return fmt.Errorf("failed to load episodes: %w", err)
	}

	if len(episodes) == 0 {
		fmt.Println("No episodes cached. Run 'podfetch fetch' first.")
		return nil
	}

	// Parse duration filters
	var minDur, maxDur time.Duration
	if minDuration != "" {
		minDur, err = time.ParseDuration(minDuration)
		if err != nil {
			return fmt.Errorf("invalid min-duration: %w", err)
		}
	}
	if maxDuration != "" {
		maxDur, err = time.ParseDuration(maxDuration)
		if err != nil {
			return fmt.Errorf("invalid max-duration: %w", err)
		}
	}

	// Filter episodes
	var filtered []store.EpisodeWithFeed
	queryLower := strings.ToLower(query)

	for _, ep := range episodes {
		// Duration filter
		if minDur > 0 && ep.Episode.Duration < minDur {
			continue
		}
		if maxDur > 0 && ep.Episode.Duration > maxDur {
			continue
		}

		// Keyword filter
		if query != "" {
			titleMatch := strings.Contains(strings.ToLower(ep.Episode.Title), queryLower)
			descMatch := strings.Contains(strings.ToLower(ep.Episode.Description), queryLower)
			if !titleMatch && !descMatch {
				continue
			}
		}

		filtered = append(filtered, ep)
	}

	// Sort by publish date (newest first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Episode.PubDate.After(filtered[j].Episode.PubDate)
	})

	// Apply limit
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	// Display results
	if len(filtered) == 0 {
		fmt.Println("No episodes match your criteria.")
		return nil
	}

	fmt.Printf("Found %d episode(s):\n\n", len(filtered))
	for _, ep := range filtered {
		printEpisode(ep)
	}

	return nil
}

func printEpisode(ep store.EpisodeWithFeed) {
	duration := feed.FormatDuration(ep.Episode.Duration)
	date := ep.Episode.PubDate.Format("2006-01-02")
	if ep.Episode.PubDate.IsZero() {
		date = "unknown"
	}

	fmt.Printf("%s %s\n", duration, ep.Episode.Title)
	fmt.Printf("      %s • %s\n\n", ep.FeedTitle, date)
}
