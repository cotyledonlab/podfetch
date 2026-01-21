# podfetch

A CLI tool to fetch and filter podcast episodes by duration. Perfect for finding episodes that fit a specific time window (e.g., a 30-minute cardio session or a 2-hour road trip).

## Installation

```bash
go install github.com/johnmaher/podfetch@latest
```

Or build from source:

```bash
git clone https://github.com/johnmaher/podfetch.git
cd podfetch
go build -o podfetch .
```

## Usage

### Add a Podcast Feed

```bash
podfetch add https://feeds.simplecast.com/l2i9YnTd
# ✓ Added: Hard Fork (150 episodes)
```

### List Tracked Feeds

```bash
podfetch list
# Tracking 2 feed(s):
#
# 1. Hard Fork
#    https://feeds.simplecast.com/l2i9YnTd
#
# 2. Lex Fridman Podcast
#    https://lexfridman.com/feed/podcast/
```

### Fetch Latest Episodes

Refresh all feeds concurrently:

```bash
podfetch fetch
# Fetching 2 feed(s) concurrently...
#
# ✓ Hard Fork (150 episodes)
# ✓ Lex Fridman Podcast (400 episodes)
#
# Fetch complete: 2 succeeded, 0 failed
```

### Search Episodes

Find episodes by duration and/or keyword:

```bash
# Episodes under 30 minutes
podfetch search --max-duration 30m

# Episodes between 1-2 hours
podfetch search --min-duration 1h --max-duration 2h

# Episodes about AI, under 45 minutes
podfetch search --query "AI" --max-duration 45m

# Limit results
podfetch search --max-duration 1h --limit 5
```

Example output:

```
Found 3 episode(s):

[32m] The Tech Episode Title
      Hard Fork • 2026-01-15

[1h 45m] Deep Dive into Neural Networks
      Lex Fridman Podcast • 2026-01-10

[28m] Quick Update on AI News
      Hard Fork • 2026-01-08
```

### Remove a Feed

Remove by name or URL:

```bash
podfetch remove "Hard Fork"
# or
podfetch remove https://feeds.simplecast.com/l2i9YnTd
```

## Sample Feeds for Testing

```bash
# Hard Fork (NYT tech podcast)
podfetch add https://feeds.simplecast.com/l2i9YnTd

# Lex Fridman Podcast
podfetch add https://lexfridman.com/feed/podcast/

# Gastropod (food science)
podfetch add https://gastropod.com/feed/
```

## Data Storage

podfetch stores data in `~/.podfetch/`:

- `feeds.json` - List of tracked feed URLs and titles
- `cache.json` - Cached episode data for offline searching

## Project Structure

```
podfetch/
├── main.go                 # Entry point
├── go.mod                  # Go module definition
├── cmd/                    # Cobra CLI commands
│   ├── root.go            # Root command setup
│   ├── add.go             # Add feed command
│   ├── list.go            # List feeds command
│   ├── fetch.go           # Fetch/refresh command
│   ├── search.go          # Search episodes command
│   └── remove.go          # Remove feed command
├── internal/
│   ├── feed/              # RSS parsing
│   │   ├── feed.go        # Feed fetching and parsing
│   │   └── feed_test.go   # Unit tests
│   └── store/             # JSON file storage
│       └── store.go       # Storage operations
└── README.md
```

## Key Go Patterns Demonstrated

### Concurrency with Goroutines + Channels

The `fetch` command uses goroutines and channels to fetch multiple feeds concurrently:

```go
// From internal/feed/feed.go
func FetchConcurrently(urls []string) <-chan FetchResult {
    results := make(chan FetchResult, len(urls))
    
    go func() {
        defer close(results)
        semaphore := make(chan struct{}, 5) // Limit concurrency
        
        for _, url := range urls {
            semaphore <- struct{}{}
            go func(feedURL string) {
                defer func() { <-semaphore }()
                feed, err := Fetch(feedURL)
                results <- FetchResult{Feed: feed, Error: err}
            }(url)
        }
        // Wait for completion...
    }()
    
    return results
}
```

### Error Handling with Sentinel Errors

```go
var (
    ErrFeedNotFound = errors.New("feed not found")
    ErrFeedExists   = errors.New("feed already exists")
)

// Usage with errors.Is()
if errors.Is(err, store.ErrFeedExists) {
    return fmt.Errorf("feed already tracked: %s", f.Title)
}
```

### Clean Package Structure

- `cmd/` - CLI layer (user interaction)
- `internal/feed/` - Domain logic (RSS parsing)
- `internal/store/` - Persistence layer (JSON storage)

## Running Tests

```bash
go test ./...
```

## License

MIT
