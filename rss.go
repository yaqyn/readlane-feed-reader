package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"time"
)

const maxFeedBytes = 8 << 20

var feedClient = &http.Client{Timeout: 30 * time.Second}

type RSSFeed struct {
	XMLName xml.Name `xml:"rss"`
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}
type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func validateFeedURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u == nil || u.Hostname() == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("feed URL must be an absolute HTTP or HTTPS URL")
	}
	if u.User != nil {
		return fmt.Errorf("feed URLs cannot contain credentials")
	}
	return nil
}
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	if err := validateFeedURL(feedURL); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")
	req.Header.Set("Accept", "application/rss+xml, application/xml, text/xml")
	resp, err := feedClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed returned HTTP %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read feed: %w", err)
	}
	if len(b) > maxFeedBytes {
		return nil, fmt.Errorf("feed exceeds %d bytes", maxFeedBytes)
	}
	var feed RSSFeed
	if err = xml.Unmarshal(b, &feed); err != nil {
		return nil, fmt.Errorf("decode RSS: %w", err)
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		item := &feed.Channel.Item[i]
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
	}
	return &feed, nil
}
func handlerFetch(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: gator fetch <url>")
	}
	feed, err := fetchFeed(s.ctx, cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Printf("%s\n%s\n", feed.Channel.Title, feed.Channel.Description)
	for _, item := range feed.Channel.Item {
		fmt.Printf("\n%s\n%s\n%s\n", item.Title, item.Link, item.Description)
	}
	return nil
}
