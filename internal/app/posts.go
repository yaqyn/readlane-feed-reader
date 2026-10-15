package app

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/yaqyn/readlane-feed-reader/internal/database"
	"strconv"
	"strings"
	"time"
)

func parsePublishedAt(raw string) sql.NullTime {
	raw = strings.TrimSpace(raw)
	for _, format := range []string{time.RFC1123Z, time.RFC822Z, time.RFC3339Nano, "Mon, 2 Jan 2006 15:04:05 -0700", "2006-01-02T15:04:05-0700", time.ANSIC, "2006-01-02"} {
		if t, err := time.Parse(format, raw); err == nil {
			return sql.NullTime{Time: t.UTC(), Valid: true}
		}
	}
	// time.Parse silently assigns offset zero to unknown zone abbreviations.
	// Recognize common RSS zones explicitly instead of misdating those posts.
	zones := map[string]int{"UT": 0, "UTC": 0, "GMT": 0, "EST": -5, "EDT": -4, "CST": -6, "CDT": -5, "MST": -7, "MDT": -6, "PST": -8, "PDT": -7}
	parts := strings.Fields(raw)
	if len(parts) > 0 {
		zone := parts[len(parts)-1]
		if offset, ok := zones[zone]; ok {
			for _, format := range []string{time.RFC1123, time.RFC822, "Mon, 2 Jan 2006 15:04:05 MST"} {
				if t, err := time.ParseInLocation(format, raw, time.FixedZone(zone, offset*3600)); err == nil {
					return sql.NullTime{Time: t.UTC(), Valid: true}
				}
			}
		}
	}
	return sql.NullTime{}
}

func savePost(s *state, feedID uuid.UUID, item RSSItem) (int64, error) {
	now := time.Now().UTC()
	return s.db.CreatePost(s.ctx, database.CreatePostParams{ID: uuid.New(), CreatedAt: now, UpdatedAt: now, Title: item.Title, Url: item.Link, Description: item.Description, PublishedAt: parsePublishedAt(item.PubDate), FeedID: feedID})
}
func handlerBrowse(s *state, cmd command, user database.User) error {
	if len(cmd.args) > 2 {
		return fmt.Errorf("usage: readlane read [limit] [offset]")
	}
	limit, offset := 5, 0
	var err error
	if len(cmd.args) > 0 {
		limit, err = strconv.Atoi(cmd.args[0])
		if err != nil || limit < 1 || limit > 100 {
			return fmt.Errorf("limit must be between 1 and 100")
		}
	}
	if len(cmd.args) > 1 {
		offset, err = strconv.Atoi(cmd.args[1])
		if err != nil || offset < 0 || offset > 1000000 {
			return fmt.Errorf("offset must be between 0 and 1000000")
		}
	}
	posts, err := s.db.GetPostsForUser(s.ctx, database.GetPostsForUserParams{UserID: user.ID, Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return databaseError("browse posts", err)
	}
	if len(posts) == 0 {
		fmt.Println("Your reading list is empty. Follow a source, then run readlane collect 1m --once.")
		return nil
	}
	fmt.Printf("%s\n%d posts for %s\n\n", cliStyle("Your reading list", "1;38;5;99"), len(posts), cleanText(user.Name, 100))
	for i, p := range posts {
		date := "Publication date unavailable"
		if p.PublishedAt.Valid {
			date = p.PublishedAt.Time.UTC().Format("02 Jan 2006, 15:04 UTC")
		}
		fmt.Printf("%s\n   %s | %s\n", cliStyle(fmt.Sprintf("%d. %s", offset+i+1, cleanText(p.Title, 300)), "1"), cleanText(p.FeedName, 200), date)
		if text := summary(p.Description); text != "" {
			fmt.Print(wrapCLI(text, 72, "   ") + "\n")
		}
		fmt.Printf("   %s\n\n", cleanText(p.Url, 4096))
	}
	return nil
}
