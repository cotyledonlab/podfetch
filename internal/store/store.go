// Package store provides JSON file-based storage for podcast feeds.
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/johnmaher/podfetch/internal/feed"
)

const (
	configDir  = ".podfetch"
	feedsFile  = "feeds.json"
	cacheFile  = "cache.json"
)

var (
	ErrFeedNotFound = errors.New("feed not found")
	ErrFeedExists   = errors.New("feed already exists")
)

// Store manages persistent storage of feeds and episodes.
type Store struct {
	dir string
	mu  sync.RWMutex
}

// FeedEntry represents a tracked feed (URL and metadata).
type FeedEntry struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

// FeedsData holds the list of tracked feeds.
type FeedsData struct {
	Feeds []FeedEntry `json:"feeds"`
}

// CacheData holds cached episode data for all feeds.
type CacheData struct {
	Feeds map[string]*feed.Feed `json:"feeds"` // keyed by URL
}

// New creates a new Store, initializing the config directory if needed.
func New() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting home directory: %w", err)
	}

	dir := filepath.Join(home, configDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating config directory: %w", err)
	}

	return &Store{dir: dir}, nil
}

// feedsPath returns the path to the feeds.json file.
func (s *Store) feedsPath() string {
	return filepath.Join(s.dir, feedsFile)
}

// cachePath returns the path to the cache.json file.
func (s *Store) cachePath() string {
	return filepath.Join(s.dir, cacheFile)
}

// LoadFeeds loads the list of tracked feeds.
func (s *Store) LoadFeeds() (*FeedsData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := &FeedsData{Feeds: []FeedEntry{}}

	content, err := os.ReadFile(s.feedsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return data, nil
		}
		return nil, fmt.Errorf("reading feeds file: %w", err)
	}

	if err := json.Unmarshal(content, data); err != nil {
		return nil, fmt.Errorf("parsing feeds file: %w", err)
	}

	return data, nil
}

// SaveFeeds saves the list of tracked feeds.
func (s *Store) SaveFeeds(data *FeedsData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding feeds: %w", err)
	}

	if err := os.WriteFile(s.feedsPath(), content, 0644); err != nil {
		return fmt.Errorf("writing feeds file: %w", err)
	}

	return nil
}

// AddFeed adds a new feed to the tracked list.
func (s *Store) AddFeed(url, title string) error {
	data, err := s.LoadFeeds()
	if err != nil {
		return err
	}

	// Check for duplicates
	for _, f := range data.Feeds {
		if f.URL == url {
			return ErrFeedExists
		}
	}

	data.Feeds = append(data.Feeds, FeedEntry{URL: url, Title: title})
	return s.SaveFeeds(data)
}

// RemoveFeed removes a feed by URL or title.
func (s *Store) RemoveFeed(identifier string) error {
	data, err := s.LoadFeeds()
	if err != nil {
		return err
	}

	found := false
	newFeeds := make([]FeedEntry, 0, len(data.Feeds))
	for _, f := range data.Feeds {
		if f.URL == identifier || f.Title == identifier {
			found = true
			continue
		}
		newFeeds = append(newFeeds, f)
	}

	if !found {
		return ErrFeedNotFound
	}

	data.Feeds = newFeeds
	return s.SaveFeeds(data)
}

// LoadCache loads the cached feed data.
func (s *Store) LoadCache() (*CacheData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data := &CacheData{Feeds: make(map[string]*feed.Feed)}

	content, err := os.ReadFile(s.cachePath())
	if err != nil {
		if os.IsNotExist(err) {
			return data, nil
		}
		return nil, fmt.Errorf("reading cache file: %w", err)
	}

	if err := json.Unmarshal(content, data); err != nil {
		return nil, fmt.Errorf("parsing cache file: %w", err)
	}

	return data, nil
}

// SaveCache saves the feed cache.
func (s *Store) SaveCache(data *CacheData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding cache: %w", err)
	}

	if err := os.WriteFile(s.cachePath(), content, 0644); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	return nil
}

// UpdateCache updates the cache for a specific feed.
func (s *Store) UpdateCache(f *feed.Feed) error {
	cache, err := s.LoadCache()
	if err != nil {
		return err
	}

	cache.Feeds[f.URL] = f
	return s.SaveCache(cache)
}

// GetAllEpisodes returns all cached episodes from all feeds.
func (s *Store) GetAllEpisodes() ([]EpisodeWithFeed, error) {
	cache, err := s.LoadCache()
	if err != nil {
		return nil, err
	}

	var episodes []EpisodeWithFeed
	for _, f := range cache.Feeds {
		for _, ep := range f.Episodes {
			episodes = append(episodes, EpisodeWithFeed{
				Episode:   ep,
				FeedTitle: f.Title,
			})
		}
	}

	return episodes, nil
}

// EpisodeWithFeed combines an episode with its feed title for display.
type EpisodeWithFeed struct {
	Episode   feed.Episode
	FeedTitle string
}
