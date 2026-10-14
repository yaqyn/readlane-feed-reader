package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	nethtml "golang.org/x/net/html"
	"strings"
	"unicode"
)

// cleanText prevents feed content from controlling the terminal and bounds output.
func cleanText(s string, max int) string {
	var b strings.Builder
	count := 0
	for _, r := range s {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			if r == '\n' || r == '\r' || r == '\t' {
				r = ' '
			} else {
				continue
			}
		}
		if count >= max {
			b.WriteString("…")
			break
		}
		b.WriteRune(r)
		count++
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
func summary(raw string) string {
	z := nethtml.NewTokenizer(strings.NewReader(raw))
	var b strings.Builder
	hidden := 0
	for {
		kind := z.Next()
		if kind == nethtml.ErrorToken {
			break
		}
		if kind == nethtml.StartTagToken || kind == nethtml.EndTagToken {
			tag, _ := z.TagName()
			name := string(tag)
			if name == "script" || name == "style" {
				if kind == nethtml.StartTagToken {
					hidden++
				} else if hidden > 0 {
					hidden--
				}
			}
			b.WriteByte(' ')
		}
		if kind == nethtml.TextToken && hidden == 0 {
			b.Write(z.Text())
		}
		if b.Len() > 8192 {
			break
		}
	}
	return cleanText(b.String(), 500)
}
func validName(s string, max int) bool {
	if strings.TrimSpace(s) == "" || len([]rune(s)) > max {
		return false
	}
	for _, r := range s {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return false
		}
	}
	return true
}
func databaseError(action string, err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%s: record not found", action)
	}
	if errors.Is(err, context.Canceled) {
		return err
	}
	var e *pq.Error
	if errors.As(err, &e) {
		switch e.Code {
		case "23505":
			return fmt.Errorf("%s: already exists", action)
		case "23503":
			return fmt.Errorf("%s: related record no longer exists", action)
		case "42P01", "42703":
			return fmt.Errorf("%s: schema is out of date; run gator migrate", action)
		}
	}
	return fmt.Errorf("%s: database operation failed", action)
}
