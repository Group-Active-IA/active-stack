package tui

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/uninstall"
)

// uninstallRunMsg carries the exit code of a RunUninstall invocation.
type uninstallRunMsg struct {
	exitCode int
}

// uninstallOutputMsg carries a line of progress output from the uninstall goroutine.
type uninstallOutputMsg struct {
	line string
}

// uninstallSelections holds the user's in-progress selections for the uninstall flow.
type uninstallSelections struct {
	agents   []model.Agent
	mode     model.InstallMode
	strategy uninstall.Strategy
}

// uninstallProgressBridge is a buffered channel that bridges io.Writer to tea.Msg.
type uninstallProgressBridge struct {
	ch chan string
}

func newUninstallBridge(n int) *uninstallProgressBridge {
	return &uninstallProgressBridge{ch: make(chan string, n)}
}

func (b *uninstallProgressBridge) Write(p []byte) (int, error) {
	lines := strings.Split(string(p), "\n")
	for _, line := range lines {
		select {
		case b.ch <- line:
		default:
		}
	}
	return len(p), nil
}

func (b *uninstallProgressBridge) close() {
	close(b.ch)
}

func listenUninstallCmd(b *uninstallProgressBridge) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-b.ch
		if !ok {
			return nil
		}
		return uninstallOutputMsg{line: line}
	}
}

// uninstallStrategyOrder defines the display order for strategy options.
var uninstallStrategyOrder = []uninstall.Strategy{
	uninstall.StrategyTargeted,
	uninstall.StrategyRestore,
}

// uninstallModeOrder defines the display order for uninstall mode options.
var uninstallModeOrder = []model.InstallMode{model.ModeLite, model.ModeFull, model.ModeCustom}

// viewUninstallFlow renders one of the sequential uninstall screens.
func (m Model) viewUninstallFlow() string {
	switch m.Screen {
	case ScreenUninstallAgents:
		return m.viewUninstallAgents()
	case ScreenUninstallMode:
		return m.viewUninstallMode()
	case ScreenUninstallStrategy:
		return m.viewUninstallStrategy()
	}
	return ""
}

// viewUninstallAgents renders the agent selection screen for uninstall.
func (m Model) viewUninstallAgents() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Uninstall — select agents") + "\n\n")
	for i, a := range m.AvailableAgents {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}
		checked := "[ ]"
		if isAgentSelected(m.uninstallSel.agents, a) {
			checked = selectedStyle.Render("[x]")
		}
		sb.WriteString(fmt.Sprintf("%s%s %s\n", cursor, checked, a))
	}
	if len(m.AvailableAgents) == 0 {
		sb.WriteString(dimStyle.Render("No agents detected.\n"))
	}
	sb.WriteString("\nSpace = toggle  Enter = continue  Esc = back\n")
	return sb.String()
}

// viewUninstallMode renders the mode selection screen for uninstall.
func (m Model) viewUninstallMode() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Uninstall — choose mode") + "\n\n")
	for i, mode := range uninstallModeOrder {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", cursor, mode))
	}
	sb.WriteString("\nEnter = select  Esc = back\n")
	return sb.String()
}

// viewUninstallStrategy renders the strategy selection screen for uninstall.
func (m Model) viewUninstallStrategy() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Uninstall — strategy") + "\n\n")
	for i, s := range uninstallStrategyOrder {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", cursor, s))
	}
	sb.WriteString("\nEnter = select  Esc = back\n")
	return sb.String()
}

// viewUninstallConfirm renders the confirmation screen for uninstall.
func (m Model) viewUninstallConfirm() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Uninstall — confirm") + "\n\n")
	sb.WriteString("Agents: ")
	if len(m.uninstallSel.agents) == 0 {
		sb.WriteString(dimStyle.Render("(all)"))
	} else {
		parts := make([]string, len(m.uninstallSel.agents))
		for i, a := range m.uninstallSel.agents {
			parts[i] = string(a)
		}
		sb.WriteString(strings.Join(parts, ", "))
	}
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Mode:   %s\n", m.uninstallSel.mode))
	sb.WriteString(fmt.Sprintf("Strategy: %s\n", m.uninstallSel.strategy))
	sb.WriteString("\n")
	sb.WriteString(errorStyle.Render("WARNING: This will modify your agent configuration.") + "\n\n")
	sb.WriteString("Enter = confirm uninstall  Esc = back\n")
	return sb.String()
}

// viewUninstalling renders the live progress screen during uninstall.
func (m Model) viewUninstalling() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Uninstalling") + "\n\n")
	for _, line := range m.uninstallLines {
		sb.WriteString(line + "\n")
	}
	if m.uninstallDone {
		sb.WriteString("\n")
		if m.uninstallExitCode == 0 {
			sb.WriteString(selectedStyle.Render("Uninstall succeeded.") + "\n")
		} else {
			sb.WriteString(errorStyle.Render("Uninstall failed.") + "\n")
		}
		sb.WriteString("\nPress Esc to return to menu.\n")
	}
	return sb.String()
}

// handleUninstallFlowKey handles key input for the sequential uninstall screens.
func (m Model) handleUninstallFlowKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.Screen {
	case ScreenUninstallAgents:
		return m.handleUninstallAgentsKey(key)
	case ScreenUninstallMode:
		return m.handleUninstallModeKey(key)
	case ScreenUninstallStrategy:
		return m.handleUninstallStrategyKey(key)
	}
	return m, nil
}

// handleUninstallAgentsKey handles key input on ScreenUninstallAgents.
func (m Model) handleUninstallAgentsKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyUp:
		if m.Cursor > 0 {
			m.Cursor--
		}
	case tea.KeyDown:
		if m.Cursor < len(m.AvailableAgents)-1 {
			m.Cursor++
		}
	case tea.KeySpace:
		if m.Cursor < len(m.AvailableAgents) {
			a := m.AvailableAgents[m.Cursor]
			m.uninstallSel.agents = toggleAgent(m.uninstallSel.agents, a)
		}
	case tea.KeyEnter:
		m.Screen = ScreenUninstallMode
		m.Cursor = 0
	case tea.KeyEsc:
		m.Screen = m.prevScreen
		m.Cursor = 0
	}
	return m, nil
}

// handleUninstallModeKey handles key input on ScreenUninstallMode.
func (m Model) handleUninstallModeKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyUp:
		if m.Cursor > 0 {
			m.Cursor--
		}
	case tea.KeyDown:
		if m.Cursor < len(uninstallModeOrder)-1 {
			m.Cursor++
		}
	case tea.KeyEnter:
		if m.Cursor < len(uninstallModeOrder) {
			m.uninstallSel.mode = uninstallModeOrder[m.Cursor]
		}
		m.Screen = ScreenUninstallStrategy
		m.Cursor = 0
	case tea.KeyEsc:
		m.Screen = ScreenUninstallAgents
		m.Cursor = 0
	}
	return m, nil
}

// handleUninstallStrategyKey handles key input on ScreenUninstallStrategy.
func (m Model) handleUninstallStrategyKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyUp:
		if m.Cursor > 0 {
			m.Cursor--
		}
	case tea.KeyDown:
		if m.Cursor < len(uninstallStrategyOrder)-1 {
			m.Cursor++
		}
	case tea.KeyEnter:
		if m.Cursor < len(uninstallStrategyOrder) {
			m.uninstallSel.strategy = uninstallStrategyOrder[m.Cursor]
		}
		m.Screen = ScreenUninstallConfirm
		m.Cursor = 0
	case tea.KeyEsc:
		m.Screen = ScreenUninstallMode
		m.Cursor = 0
	}
	return m, nil
}

// handleUninstallConfirmKey handles key input on ScreenUninstallConfirm.
func (m Model) handleUninstallConfirmKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyEnter:
		// Build ParsedUninstallFlags from selections and launch goroutine.
		flags := headless.ParsedUninstallFlags{
			DryRun: false,
			Yes:    true,
			Intent: uninstall.Intent{
				Mode:     m.uninstallSel.mode,
				Agents:   m.uninstallSel.agents,
				Strategy: m.uninstallSel.strategy,
			},
		}
		m.Screen = ScreenUninstalling
		m.uninstallLines = nil
		m.uninstallDone = false
		m.uninstallExitCode = 0

		if m.deps.RunUninstall == nil {
			m.uninstallDone = true
			m.uninstallExitCode = 1
			return m, nil
		}

		bridge := newUninstallBridge(128)
		m.uninstallBridge = bridge
		runFn := m.deps.RunUninstall
		go func() {
			defer bridge.close()
			code := runFn(flags, bridge)
			bridge.ch <- fmt.Sprintf("__exitcode__%d", code)
		}()
		return m, listenUninstallCmd(bridge)

	case tea.KeyEsc:
		m.Screen = ScreenUninstallStrategy
		m.Cursor = 0
	}
	return m, nil
}

// handleUninstallingKey handles key input on ScreenUninstalling.
func (m Model) handleUninstallingKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.uninstallDone && key.Type == tea.KeyEsc {
		m.Screen = m.prevScreen
		m.Cursor = 0
	}
	return m, nil
}

// Update handles uninstall-specific messages.
// Called from the main Update before the default switch.
func (m Model) applyUninstallMsg(msg tea.Msg) (Model, tea.Cmd, bool) {
	switch msg := msg.(type) {
	case uninstallOutputMsg:
		line := msg.line
		if strings.HasPrefix(line, "__exitcode__") {
			var code int
			fmt.Sscanf(line, "__exitcode__%d", &code)
			m.uninstallExitCode = code
			m.uninstallDone = true
			return m, nil, true
		}
		if line != "" {
			m.uninstallLines = append(m.uninstallLines, line)
		}
		if m.uninstallBridge != nil {
			return m, listenUninstallCmd(m.uninstallBridge), true
		}
		return m, nil, true
	}
	return m, nil, false
}

// io.Writer type for uninstall bridge (satisfies io.Writer interface)
var _ io.Writer = (*uninstallProgressBridge)(nil)
