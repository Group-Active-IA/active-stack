package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// starterRunMsg carries the exit code of a RunStarter invocation.
type starterRunMsg struct {
	exitCode int
}

// enterStarters transitions the model to ScreenStarters.
func (m Model) enterStarters() (tea.Model, tea.Cmd) {
	m.Screen = ScreenStarters
	m.Cursor = 0
	return m, nil
}

// viewStarters renders the starters list screen.
func (m Model) viewStarters() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Starters") + "\n\n")
	if len(m.deps.Starters) == 0 {
		sb.WriteString(dimStyle.Render("No starters available.") + "\n")
		sb.WriteString("\nEsc = back\n")
		return sb.String()
	}
	for i, s := range m.deps.Starters {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}
		desc := s.Description
		if desc != "" {
			sb.WriteString(fmt.Sprintf("%s%s — %s\n", cursor, s.Name, desc))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s\n", cursor, s.Name))
		}
	}
	sb.WriteString("\nEnter = install into current dir  Esc = back\n")
	return sb.String()
}

// handleStartersKey handles keyboard input on ScreenStarters.
func (m Model) handleStartersKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyUp:
		if m.Cursor > 0 {
			m.Cursor--
		}
	case tea.KeyDown:
		if m.Cursor < len(m.deps.Starters)-1 {
			m.Cursor++
		}
	case tea.KeyEnter:
		if len(m.deps.Starters) == 0 {
			break
		}
		if m.Cursor >= len(m.deps.Starters) {
			break
		}
		if m.deps.RunStarter == nil {
			break
		}
		starterID := m.deps.Starters[m.Cursor].ID
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		runFn := m.deps.RunStarter
		var w io.Writer = io.Discard
		return m, func() tea.Msg {
			code := runFn(starterID, cwd, m.deps.AvailableAgents, w)
			return starterRunMsg{exitCode: code}
		}
	case tea.KeyEsc:
		m.Screen = m.prevScreen
		m.Cursor = 0
	}
	return m, nil
}
