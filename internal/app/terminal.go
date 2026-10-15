package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/google/uuid"
	"github.com/muesli/termenv"
	"github.com/yaqyn/readlane-feed-reader/internal/database"
)

const terminalPageSize = 20

type terminalData struct {
	seq     int
	posts   []database.GetReaderPostsRow
	sources []database.GetReaderSourcesRow
	total   int64
	page    int
	err     error
}
type searchDue struct{ seq int }
type terminalModel struct {
	state                                              *state
	user                                               database.User
	width, height, selected, sourceSelected, page, seq int
	mode                                               string
	source                                             uuid.NullUUID
	sourceName                                         string
	posts                                              []database.GetReaderPostsRow
	sources                                            []database.GetReaderSourcesRow
	total                                              int64
	input                                              textinput.Model
	searching, loading                                 bool
	err                                                error
	detail                                             viewport.Model
}

func newTerminalModel(s *state, user database.User) terminalModel {
	input := textinput.New()
	input.Placeholder = "Search your reading list"
	input.CharLimit = 200
	input.Prompt = "/ "
	input.Width = 48
	return terminalModel{state: s, user: user, width: 80, height: 24, mode: "reading", page: 1, input: input, loading: true, detail: viewport.New(72, 12)}
}
func terminalAvailable() bool {
	in, e := os.Stdin.Stat()
	out, f := os.Stdout.Stat()
	return e == nil && f == nil && in.Mode()&os.ModeCharDevice != 0 && out.Mode()&os.ModeCharDevice != 0 && os.Getenv("TERM") != "dumb"
}
func handlerTerminal(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 0 {
		return fmt.Errorf("usage: readlane tui")
	}
	if !terminalAvailable() {
		return fmt.Errorf("the interactive reader needs a terminal; use readlane read for plain output")
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		lipgloss.SetColorProfile(termenv.Ascii)
	}
	_, err := tea.NewProgram(newTerminalModel(s, user), tea.WithAltScreen(), tea.WithContext(s.ctx)).Run()
	if s.ctx.Err() != nil || errors.Is(err, tea.ErrProgramKilled) {
		return nil
	}
	return err
}
func (m terminalModel) Init() tea.Cmd { return m.load() }
func (m terminalModel) load() tea.Cmd {
	return func() tea.Msg {
		result := terminalData{seq: m.seq, page: m.page}
		ctx, cancel := context.WithTimeout(m.state.ctx, 5*time.Second)
		defer cancel()
		q := m.state.db
		result.total, result.err = q.CountReaderPosts(ctx, database.CountReaderPostsParams{UserID: m.user.ID, SourceID: m.source, Search: m.input.Value()})
		if result.err != nil {
			return result
		}
		result.page = min(max(1, m.page), max(1, int((result.total+terminalPageSize-1)/terminalPageSize)))
		result.posts, result.err = q.GetReaderPosts(ctx, database.GetReaderPostsParams{UserID: m.user.ID, SourceID: m.source, Search: m.input.Value(), PageOffset: int32((result.page - 1) * terminalPageSize), PageLimit: terminalPageSize})
		if result.err != nil {
			return result
		}
		result.sources, result.err = q.GetReaderSources(ctx, m.user.ID)
		return result
	}
}
func (m *terminalModel) reload() tea.Cmd { m.seq++; m.loading = true; m.err = nil; return m.load() }
func (m *terminalModel) prepareDetail() {
	if m.selected >= len(m.posts) {
		return
	}
	post := m.posts[m.selected]
	width := max(20, min(78, m.width-10))
	m.detail.Width = max(20, m.width-8)
	m.detail.Height = max(3, m.height-11)
	date := "Publication date unavailable"
	if post.PublishedAt.Valid {
		date = post.PublishedAt.Time.UTC().Format("02 January 2006, 15:04 UTC")
	}
	title := lipgloss.NewStyle().Bold(true).Width(width).Render(cleanText(post.Title, 300))
	content := title + "\n\n" + cleanText(post.FeedName, 200) + "\n" + date + "\n\n" + "Feed summary\n\n" + ansi.Wrap(plainText(post.Description, 12000), width, "") + "\n\nOriginal article\n" + ansi.Wrap(cleanText(post.Url, 4096), width, "")
	m.detail.SetContent(content)
	m.detail.GotoTop()
}
func (m terminalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		oldOffset := m.detail.YOffset
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = max(10, min(60, msg.Width-10))
		if m.mode == "detail" {
			m.prepareDetail()
			m.detail.SetYOffset(oldOffset)
		}
		return m, nil
	case terminalData:
		if msg.seq != m.seq {
			return m, nil
		}
		m.loading = false
		if msg.err != nil {
			m.err = databaseError("load reading list", msg.err)
			return m, nil
		}
		m.posts = msg.posts
		m.sources = msg.sources
		if len(m.sources) > 200 {
			m.sources = m.sources[:200]
		}
		m.total = msg.total
		m.page = msg.page
		m.selected = min(m.selected, max(0, len(m.posts)-1))
		m.sourceSelected = min(m.sourceSelected, max(0, len(m.sources)))
		return m, nil
	case searchDue:
		if msg.seq != m.seq {
			return m, nil
		}
		return m, m.load()
	case tea.KeyMsg:
		key := msg.String()
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		if m.searching {
			switch key {
			case "esc":
				m.searching = false
				m.input.Blur()
				return m, nil
			case "enter":
				m.searching = false
				m.input.Blur()
				m.page = 1
				m.selected = 0
				return m, m.reload()
			}
			before := m.input.Value()
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			if before != m.input.Value() {
				m.page = 1
				m.selected = 0
				m.seq++
				m.loading = true
				seq := m.seq
				return m, tea.Batch(cmd, tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg { return searchDue{seq: seq} }))
			}
			return m, cmd
		}
		if key == "q" {
			return m, tea.Quit
		}
		if m.mode == "detail" {
			if key == "esc" || key == "backspace" || key == "left" {
				m.mode = "reading"
				return m, nil
			}
			var cmd tea.Cmd
			m.detail, cmd = m.detail.Update(msg)
			return m, cmd
		}
		switch key {
		case "/":
			m.mode = "reading"
			m.searching = true
			return m, m.input.Focus()
		case "tab":
			if m.mode == "sources" {
				m.mode = "reading"
			} else {
				m.mode = "sources"
			}
			return m, nil
		case "r":
			return m, m.reload()
		case "ctrl+u":
			m.input.SetValue("")
			m.page = 1
			m.selected = 0
			return m, m.reload()
		case "a":
			m.source = uuid.NullUUID{}
			m.sourceName = ""
			m.page = 1
			m.mode = "reading"
			return m, m.reload()
		case "down", "j":
			if m.mode == "sources" {
				m.sourceSelected = min(m.sourceSelected+1, len(m.sources))
			} else {
				m.selected = min(m.selected+1, max(0, len(m.posts)-1))
			}
		case "up", "k":
			if m.mode == "sources" {
				m.sourceSelected = max(m.sourceSelected-1, 0)
			} else {
				m.selected = max(m.selected-1, 0)
			}
		case "right", "n", "pgdown":
			if m.mode == "reading" && int64(m.page*terminalPageSize) < m.total && !m.loading {
				m.page++
				m.selected = 0
				return m, m.reload()
			}
		case "left", "p", "pgup":
			if m.mode == "reading" && m.page > 1 && !m.loading {
				m.page--
				m.selected = 0
				return m, m.reload()
			}
		case "enter":
			if m.mode == "sources" {
				m.source = uuid.NullUUID{}
				m.sourceName = ""
				if m.sourceSelected > 0 {
					source := m.sources[m.sourceSelected-1]
					m.source = uuid.NullUUID{UUID: source.ID, Valid: true}
					m.sourceName = source.Name
				}
				m.mode = "reading"
				m.page = 1
				m.selected = 0
				return m, m.reload()
			}
			if len(m.posts) > 0 && !m.loading {
				m.mode = "detail"
				m.prepareDetail()
			}
		}
	}
	if m.searching {
		before := m.input.Value()
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		if before != m.input.Value() {
			m.page = 1
			m.selected = 0
			m.seq++
			m.loading = true
			seq := m.seq
			return m, tea.Batch(cmd, tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg { return searchDue{seq: seq} }))
		}
		return m, cmd
	}
	return m, nil
}
func (m terminalModel) View() string {
	width := max(16, m.width-4)
	height := max(4, m.height-9)
	accent := lipgloss.AdaptiveColor{Light: brandPrimary, Dark: brandTerminalPrimaryDark}
	ink := lipgloss.AdaptiveColor{Light: brandInk, Dark: brandTerminalInkDark}
	muted := lipgloss.AdaptiveColor{Light: brandMuted, Dark: brandTerminalMutedDark}
	tint := lipgloss.AdaptiveColor{Light: brandTint, Dark: brandTerminalSelectedDark}
	brand := lipgloss.NewStyle().Foreground(accent).Bold(true)
	quiet := lipgloss.NewStyle().Foreground(muted)
	bodyStyle := lipgloss.NewStyle().Width(width).Height(height).MaxHeight(height).Foreground(ink)
	if m.width < 40 || m.height < 14 {
		return "\n  Readlane\n\n  Enlarge to at least 40 columns and 14 rows,\n  or use readlane read for plain output.\n\n  q quit\n"
	}
	header := brand.Render("▌ readlane") + quiet.Render("   Your feeds. Your pace.")
	reader := quiet.Render("Reading as ") + cleanText(m.user.Name, 100)
	if lipgloss.Width(header)+lipgloss.Width(reader)+3 < width {
		header += strings.Repeat(" ", width-lipgloss.Width(header)-lipgloss.Width(reader)) + reader
	}
	readingTab, sourceTab := "Reading list", "Sources"
	if m.mode == "sources" {
		sourceTab = brand.Render("Sources")
	} else {
		readingTab = brand.Render("Reading list")
	}
	nav := readingTab + "    " + sourceTab + quiet.Render("    tab to switch")
	noun := "posts"
	if m.total == 1 {
		noun = "post"
	}
	caption := fmt.Sprintf("%d %s  Page %d/%d", m.total, noun, m.page, max(1, int((m.total+terminalPageSize-1)/terminalPageSize)))
	if m.source.Valid {
		caption += " from " + cleanText(m.sourceName, 100)
	}
	if m.input.Value() != "" {
		caption += " matching “" + cleanText(m.input.Value(), 200) + "”"
	}
	if m.mode == "sources" {
		caption = fmt.Sprintf("%d followed sources", len(m.sources))
		if len(m.sources) == 200 {
			caption += " shown; the CLI lists the complete registry"
		}
	}
	if m.loading {
		caption = "Gathering your reading list…"
	}
	if m.err != nil {
		caption = "Could not refresh. Press r to try again."
	}
	var body strings.Builder
	if m.mode == "detail" {
		body.WriteString(m.detail.View())
	} else if m.mode == "sources" {
		available := max(1, height/2)
		start := max(0, m.sourceSelected-available+1)
		end := min(len(m.sources)+1, start+available)
		for i := start; i < end; i++ {
			name := "All followed sources"
			if i > 0 {
				name = cleanText(m.sources[i-1].Name, 200)
			}
			line := ansi.Truncate(name, width-6, "…")
			prefix := "  "
			if i == m.sourceSelected {
				prefix = "› "
				line = lipgloss.NewStyle().Bold(true).Foreground(accent).Render(line)
			}
			body.WriteString(prefix + line + "\n\n")
		}
		if len(m.sources) == 0 {
			body.WriteString(quiet.Render("Add a source with readlane addfeed <name> <url>."))
		}
		if len(m.sources) > 200 {
			body.WriteString("\nFirst 200 sources shown; use readlane sources for the full registry.")
		}
	} else if len(m.posts) == 0 {
		if m.loading {
			body.WriteString("Your stories will appear here.")
		} else if m.err != nil {
			body.WriteString("Check PostgreSQL and run readlane doctor.\n\nYour filters are preserved; press r to retry.")
		} else if m.input.Value() != "" {
			body.WriteString("No stories match this search.\n\nPress ctrl+u to clear it, or / to try another phrase.")
		} else {
			body.WriteString("Make room for what you follow.\n\nAdd a source:\n  readlane addfeed \"Lane's Blog\" https://www.wagslane.dev/index.xml\n\nCollect a first batch:\n  readlane collect 1m --once")
		}
	} else {
		linesPerPost := 3
		if m.width >= 90 {
			linesPerPost = 4
		}
		visible := max(1, height/linesPerPost)
		start := max(0, m.selected-visible+1)
		end := min(len(m.posts), start+visible)
		for i := start; i < end; i++ {
			p := m.posts[i]
			date := "Undated"
			if p.PublishedAt.Valid {
				date = p.PublishedAt.Time.UTC().Format("02 Jan 2006")
			}
			title := ansi.Truncate(cleanText(p.Title, 300), width-5, "…")
			meta := ansi.Truncate(cleanText(p.FeedName, 200)+"  ·  "+date, width-5, "…")
			prefix := "  "
			lineStyle := lipgloss.NewStyle().Width(width).Padding(0, 1)
			if i == m.selected {
				prefix = "› "
				lineStyle = lineStyle.Background(tint)
				title = brand.Render(title)
			}
			block := prefix + title + "\n  " + quiet.Render(meta)
			if linesPerPost == 4 {
				block += "\n  " + quiet.Render(ansi.Truncate(summary(p.Description), width-5, "…"))
			}
			body.WriteString(lineStyle.Render(block) + "\n\n")
		}
	}
	search := quiet.Render("/ search")
	if m.searching {
		input := m.input
		input.Width = max(8, width-4)
		search = ansi.Truncate(input.View(), width, "…")
	}
	footer := "↑/↓ move   enter read   / search   tab sources   n/p page   r refresh   q quit"
	if m.mode == "detail" {
		footer = "↑/↓ scroll   pgup/pgdown page   esc back   q quit"
	}
	if m.mode == "sources" {
		footer = "↑/↓ choose   enter filter   a all sources   tab reading list   q quit"
	}
	if m.searching {
		footer = "Type to search   ctrl+u clear   enter apply   esc close"
	}
	if width < 100 {
		footer = "↑↓ move  enter read  / search  tab sources  n/p pages  q quit"
		if width < 60 {
			footer = "↑↓ move  enter read  / find  q quit"
		}
		if m.mode == "sources" {
			footer = "↑↓ choose  enter filter  tab reading  a all  q quit"
			if width < 60 {
				footer = "↑↓ choose  enter filter  q quit"
			}
		}
		if m.mode == "detail" {
			footer = "↑↓ scroll  esc back  q quit"
		}
		if m.searching {
			footer = "enter apply  esc close  ctrl+u clear"
		}
	}
	pagination := fmt.Sprintf("Page %d of %d", m.page, max(1, int((m.total+terminalPageSize-1)/terminalPageSize)))
	footerText := pagination + "    " + footer
	if width < 100 {
		footerText = footer
	}
	return lipgloss.NewStyle().Padding(1, 2).Render(ansi.Truncate(header, width, "…") + "\n\n" + ansi.Truncate(nav, width, "…") + "\n" + quiet.Render(ansi.Truncate(caption, width, "…")) + "\n\n" + bodyStyle.Render(body.String()) + "\n" + search + "\n" + quiet.Render(ansi.Truncate(footerText, width, "…")))
}
