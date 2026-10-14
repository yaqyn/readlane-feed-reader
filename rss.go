package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxFeedBytes = 8 << 20

var feedClient = &http.Client{
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return fmt.Errorf("too many feed redirects")
		}
		if err := validateFeedURL(req.URL.String()); err != nil {
			return err
		}
		if len(via) > 0 && via[0].URL.Scheme == "https" && req.URL.Scheme != "https" {
			return fmt.Errorf("refusing HTTPS to HTTP redirect")
		}
		return nil
	},
}

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
	if len(raw) > 4096 {
		return fmt.Errorf("feed URL exceeds 4096 bytes")
	}
	for _, r := range raw {
		if r < 32 || r == 127 {
			return fmt.Errorf("feed URL contains control characters")
		}
	}
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
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
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
	return parseFeed(b, resp.Request.URL)
}

// Atom is normalized to the same item model as RSS.
type atomFeed struct {
	XMLName  xml.Name `xml:"feed"`
	Title    string   `xml:"title"`
	Subtitle string   `xml:"subtitle"`
	Entries  []struct {
		Title string `xml:"title"`
		Links []struct {
			Href string `xml:"href,attr"`
			Rel  string `xml:"rel,attr"`
		} `xml:"link"`
		Summary   string `xml:"summary"`
		Content   string `xml:"content"`
		Published string `xml:"published"`
		Updated   string `xml:"updated"`
	} `xml:"entry"`
}

func parseFeed(b []byte, base *url.URL) (*RSSFeed, error) {
	var root struct{ XMLName xml.Name }
	if err := xml.Unmarshal(b, &root); err != nil {
		return nil, fmt.Errorf("decode feed: malformed XML")
	}
	var feed RSSFeed
	switch root.XMLName.Local {
	case "rss":
		if err := xml.Unmarshal(b, &feed); err != nil {
			return nil, fmt.Errorf("decode RSS: malformed XML")
		}
	case "feed":
		var atom atomFeed
		if err := xml.Unmarshal(b, &atom); err != nil {
			return nil, fmt.Errorf("decode Atom: malformed XML")
		}
		feed.Channel.Title = atom.Title
		feed.Channel.Description = atom.Subtitle
		for _, entry := range atom.Entries {
			link := ""
			for _, l := range entry.Links {
				if l.Rel == "" || l.Rel == "alternate" {
					link = l.Href
					break
				}
			}
			description := entry.Summary
			if description == "" {
				description = entry.Content
			}
			date := entry.Published
			if date == "" {
				date = entry.Updated
			}
			feed.Channel.Item = append(feed.Channel.Item, RSSItem{Title: entry.Title, Link: link, Description: description, PubDate: date})
		}
	default:
		return nil, fmt.Errorf("unsupported feed format: expected RSS or Atom")
	}
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		item := &feed.Channel.Item[i]
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		item.PubDate = strings.TrimSpace(item.PubDate)
		if raw := strings.TrimSpace(item.Link); raw != "" {
			if u, err := url.Parse(raw); err == nil {
				if base != nil {
					u = base.ResolveReference(u)
				}
				u.Fragment = ""
				item.Link = u.String()
			}
		}
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
	fmt.Printf("%s\n%s\n", cleanText(feed.Channel.Title, 300), summary(feed.Channel.Description))
	for _, item := range feed.Channel.Item {
		fmt.Printf("\n%s\n%s\n%s\n", cleanText(item.Title, 300), cleanText(item.Link, 4096), summary(item.Description))
	}
	return nil
}
