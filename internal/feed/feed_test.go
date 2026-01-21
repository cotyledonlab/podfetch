package feed

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{
			name:     "HH:MM:SS format",
			input:    "01:30:45",
			expected: 1*time.Hour + 30*time.Minute + 45*time.Second,
		},
		{
			name:     "MM:SS format",
			input:    "45:30",
			expected: 45*time.Minute + 30*time.Second,
		},
		{
			name:     "seconds only",
			input:    "3600",
			expected: 1 * time.Hour,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "whitespace",
			input:    "  30:00  ",
			expected: 30 * time.Minute,
		},
		{
			name:     "hours with leading zeros",
			input:    "02:15:30",
			expected: 2*time.Hour + 15*time.Minute + 30*time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDuration(tt.input)
			if result != tt.expected {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
	}{
		{
			name:     "RFC1123Z format",
			input:    "Mon, 02 Jan 2006 15:04:05 -0700",
			expected: time.Date(2006, 1, 2, 15, 4, 5, 0, time.FixedZone("", -7*3600)),
		},
		{
			name:     "empty string",
			input:    "",
			expected: time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDate(tt.input)
			if !result.Equal(tt.expected) {
				t.Errorf("parseDate(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple tags",
			input:    "<p>Hello</p>",
			expected: "Hello",
		},
		{
			name:     "nested tags",
			input:    "<div><p>Hello <strong>World</strong></p></div>",
			expected: "Hello World",
		},
		{
			name:     "no tags",
			input:    "Plain text",
			expected: "Plain text",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripHTML(tt.input)
			if result != tt.expected {
				t.Errorf("stripHTML(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Duration
		expected string
	}{
		{
			name:     "minutes only",
			input:    32 * time.Minute,
			expected: "[32m]",
		},
		{
			name:     "hours and minutes",
			input:    2*time.Hour + 15*time.Minute,
			expected: "[2h 15m]",
		},
		{
			name:     "zero duration",
			input:    0,
			expected: "[??m]",
		},
		{
			name:     "one hour",
			input:    1 * time.Hour,
			expected: "[1h 00m]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.input)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParse(t *testing.T) {
	rssData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Podcast</title>
    <description>A test podcast feed</description>
    <item>
      <title>Episode 1</title>
      <description>First episode description</description>
      <pubDate>Mon, 15 Jan 2024 10:00:00 +0000</pubDate>
      <duration>30:00</duration>
      <guid>ep1</guid>
      <enclosure url="https://example.com/ep1.mp3" type="audio/mpeg"/>
    </item>
    <item>
      <title>Episode 2</title>
      <description><![CDATA[<p>Second episode with <strong>HTML</strong></p>]]></description>
      <pubDate>Mon, 22 Jan 2024 10:00:00 +0000</pubDate>
      <duration>01:15:30</duration>
      <guid>ep2</guid>
      <enclosure url="https://example.com/ep2.mp3" type="audio/mpeg"/>
    </item>
  </channel>
</rss>`)

	feed, err := Parse(rssData)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if feed.Title != "Test Podcast" {
		t.Errorf("feed.Title = %q, want %q", feed.Title, "Test Podcast")
	}

	if len(feed.Episodes) != 2 {
		t.Fatalf("len(feed.Episodes) = %d, want 2", len(feed.Episodes))
	}

	ep1 := feed.Episodes[0]
	if ep1.Title != "Episode 1" {
		t.Errorf("ep1.Title = %q, want %q", ep1.Title, "Episode 1")
	}
	if ep1.Duration != 30*time.Minute {
		t.Errorf("ep1.Duration = %v, want %v", ep1.Duration, 30*time.Minute)
	}

	ep2 := feed.Episodes[1]
	expectedDuration := 1*time.Hour + 15*time.Minute + 30*time.Second
	if ep2.Duration != expectedDuration {
		t.Errorf("ep2.Duration = %v, want %v", ep2.Duration, expectedDuration)
	}
}
