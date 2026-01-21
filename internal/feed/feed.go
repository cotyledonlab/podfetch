// Package feed provides RSS/Atom feed parsing functionality.
package feed

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Episode represents a single podcast episode.
type Episode struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	PubDate     time.Time     `json:"pub_date"`
	Duration    time.Duration `json:"duration"`
	AudioURL    string        `json:"audio_url"`
	GUID        string        `json:"guid"`
}

// Feed represents a podcast feed with its episodes.
type Feed struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	LastFetched time.Time `json:"last_fetched"`
	Episodes    []Episode `json:"episodes"`
}

// RSS XML structures for parsing
type rss struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string    `xml:"title"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string       `xml:"title"`
	Description string       `xml:"description"`
	PubDate     string       `xml:"pubDate"`
	Duration    string       `xml:"duration"`
	GUID        string       `xml:"guid"`
	Enclosure   rssEnclosure `xml:"enclosure"`
}

type rssEnclosure struct {
	URL  string `xml:"url,attr"`
	Type string `xml:"type,attr"`
}

// FetchResult is returned through channels when fetching feeds concurrently.
type FetchResult struct {
	Feed  *Feed
	Error error
}

// Fetch retrieves and parses a podcast feed from the given URL.
func Fetch(url string) (*Feed, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	// Set a user-agent to avoid being blocked
	req.Header.Set("User-Agent", "podfetch/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	feed, err := Parse(body)
	if err != nil {
		return nil, err
	}

	feed.URL = url
	feed.LastFetched = time.Now()

	return feed, nil
}

// Parse parses RSS XML data into a Feed struct.
func Parse(data []byte) (*Feed, error) {
	var rss rss
	if err := xml.Unmarshal(data, &rss); err != nil {
		return nil, fmt.Errorf("parsing RSS: %w", err)
	}

	feed := &Feed{
		Title:       strings.TrimSpace(rss.Channel.Title),
		Description: strings.TrimSpace(rss.Channel.Description),
		Episodes:    make([]Episode, 0, len(rss.Channel.Items)),
	}

	for _, item := range rss.Channel.Items {
		episode := Episode{
			Title:       strings.TrimSpace(item.Title),
			Description: stripHTML(strings.TrimSpace(item.Description)),
			PubDate:     parseDate(item.PubDate),
			Duration:    parseDuration(item.Duration),
			AudioURL:    item.Enclosure.URL,
			GUID:        item.GUID,
		}
		feed.Episodes = append(feed.Episodes, episode)
	}

	return feed, nil
}

// parseDuration handles various duration formats:
// - "HH:MM:SS" or "MM:SS"
// - "3600" (seconds as string)
// - "1h30m" (Go duration format)
func parseDuration(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// Try HH:MM:SS or MM:SS format
	if strings.Contains(s, ":") {
		parts := strings.Split(s, ":")
		var hours, minutes, seconds int

		switch len(parts) {
		case 3:
			hours, _ = strconv.Atoi(parts[0])
			minutes, _ = strconv.Atoi(parts[1])
			seconds, _ = strconv.Atoi(parts[2])
		case 2:
			minutes, _ = strconv.Atoi(parts[0])
			seconds, _ = strconv.Atoi(parts[1])
		}

		return time.Duration(hours)*time.Hour +
			time.Duration(minutes)*time.Minute +
			time.Duration(seconds)*time.Second
	}

	// Try plain seconds
	if secs, err := strconv.Atoi(s); err == nil {
		return time.Duration(secs) * time.Second
	}

	// Try Go duration format (e.g., "1h30m")
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}

	return 0
}

// parseDate attempts to parse various date formats used in RSS feeds.
func parseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}

	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t
		}
	}

	return time.Time{}
}

// stripHTML removes HTML tags from a string.
func stripHTML(s string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(s, "")
}

// FormatDuration formats a duration as "[XXm]" or "[Xh XXm]".
func FormatDuration(d time.Duration) string {
	if d == 0 {
		return "[??m]"
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("[%dh %02dm]", hours, minutes)
	}
	return fmt.Sprintf("[%dm]", minutes)
}

// FetchConcurrently fetches multiple feeds concurrently using goroutines.
// It returns a channel that will receive FetchResult for each feed.
func FetchConcurrently(urls []string) <-chan FetchResult {
	results := make(chan FetchResult, len(urls))

	go func() {
		defer close(results)

		// Create a channel to limit concurrency
		semaphore := make(chan struct{}, 5) // Max 5 concurrent fetches

		for _, url := range urls {
			semaphore <- struct{}{} // Acquire

			go func(feedURL string) {
				defer func() { <-semaphore }() // Release

				feed, err := Fetch(feedURL)
				results <- FetchResult{Feed: feed, Error: err}
			}(url)
		}

		// Wait for all goroutines to finish
		for i := 0; i < cap(semaphore); i++ {
			semaphore <- struct{}{}
		}
	}()

	return results
}
