package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/google/uuid"
	"github.com/yaqyn/readlane/internal/database"
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
