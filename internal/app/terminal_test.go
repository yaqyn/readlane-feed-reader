package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/google/uuid"
	"github.com/yaqyn/readlane-feed-reader/internal/database"
	"strings"
	"testing"
)

func terminalKey(m terminalModel, key tea.KeyMsg) terminalModel {
	model, _ := m.Update(key)
	return model.(terminalModel)
}
func TestTerminalNavigationSearchAndResponsiveLayout(t *testing.T) {
	m := newTerminalModel(nil, database.User{Name: "alice"})
	m.loading = false
	m.width = 100
	m.height = 30
	m.total = 2
	m.posts = []database.GetReaderPostsRow{{ID: uuid.New(), Title: "First story", FeedName: "Example", Description: "Readable summary"}, {ID: uuid.New(), Title: "Second story", FeedName: "Example"}}
	m = terminalKey(m, tea.KeyMsg{Type: tea.KeyDown})
	if m.selected != 1 {
		t.Fatal("down did not select next post")
	}
	m = terminalKey(m, tea.KeyMsg{Type: tea.KeyEnter})
	if m.mode != "detail" || !strings.Contains(m.detail.View(), "Second story") {
		t.Fatal("enter did not open summary")
	}
	m = terminalKey(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.mode != "reading" {
		t.Fatal("escape did not return to list")
	}
	m = terminalKey(m, tea.KeyMsg{Type: tea.KeyTab})
	if m.mode != "sources" {
		t.Fatal("tab did not switch views")
	}
	m = terminalKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !m.searching {
		t.Fatal("search did not focus")
	}
	m = terminalKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if m.input.Value() != "q" {
		t.Fatal("q quit rather than typing in search")
	}
	m = terminalKey(m, tea.KeyMsg{Type: tea.KeyCtrlU})
	if m.input.Value() != "" {
		t.Fatal("search clear failed")
	}
	for _, width := range []int{40, 80, 120} {
		m.width = width
		for _, line := range strings.Split(m.View(), "\n") {
			if ansi.StringWidth(line) > width {
				t.Fatalf("terminal overflow at width %d: %q", width, line)
			}
		}
	}
	m.width = 30
	if !strings.Contains(m.View(), "Enlarge") {
		t.Fatal("small terminal has no recovery guidance")
	}
}
func TestTerminalIgnoresStaleSearchResults(t *testing.T) {
	m := newTerminalModel(nil, database.User{})
	m.seq = 2
	result, _ := m.Update(terminalData{seq: 1, total: 99})
	if result.(terminalModel).total != 0 {
		t.Fatal("stale results replaced current state")
	}
}

func TestTerminalControlsMotionAndScreenBounds(t *testing.T) {
	m := newTerminalModel(nil, database.User{Name: "reader"})
	m.loading = false
	m.posts = []database.GetReaderPostsRow{{Title: "A long story about reading", FeedName: "Example", Description: strings.Repeat("A readable paragraph. ", 100)}}
	m = terminalKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m.help || !strings.Contains(m.View(), "controls") {
		t.Fatal("help did not open")
	}
	m = terminalKey(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.help {
		t.Fatal("escape did not close help")
	}
	m = terminalKey(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	if !m.reducedMotion || m.pulse() != nil {
		t.Fatal("reduced motion still schedules animation")
	}
	m.loading = true
	m.reducedMotion = false
	next, cmd := m.Update(terminalPulse{})
	if next.(terminalModel).frame != 1 || cmd == nil {
		t.Fatal("loading does not animate")
	}
	m.loading = false
	_, cmd = m.Update(terminalPulse{})
	if cmd != nil {
		t.Fatal("idle screen schedules animation")
	}
	for _, size := range [][2]int{{40, 14}, {60, 20}, {80, 24}, {120, 40}} {
		for _, mode := range []string{"reading", "sources", "detail"} {
			m.width, m.height, m.mode = size[0], size[1], mode
			m.prepareDetail()
			for _, help := range []bool{false, true} {
				m.help = help
				view := m.View()
				if lipgloss.Height(view) > m.height {
					t.Fatalf("%s exceeds height %d", mode, m.height)
				}
				for _, line := range strings.Split(view, "\n") {
					if ansi.StringWidth(line) > m.width {
						t.Fatalf("%s exceeds width %d", mode, m.width)
					}
				}
			}
		}
	}
}
