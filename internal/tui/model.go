package tui

import (
	"fmt"
	"io"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/JuanCruzRobledo/jr-stack/cmd/jr-stack/headless"
	"github.com/JuanCruzRobledo/jr-stack/internal/install"
	"github.com/JuanCruzRobledo/jr-stack/internal/model"
	"github.com/JuanCruzRobledo/jr-stack/internal/pipeline"
	"github.com/JuanCruzRobledo/jr-stack/internal/system"
)

// ModelDeps holds the injected dependencies for the TUI model.
// All fields are interfaces / plain values so tests can inject fakes.
type ModelDeps struct {
	// Catalog is the harness catalog (install.Catalog interface).
	Catalog install.Catalog
	// Registry maps agents to their adapters (install.Registry interface).
	Registry install.Registry
	// HomeDir is the user home directory passed to adapters.
	HomeDir string
	// AvailableAgents is the intersection of detected + registered agents
	// (pre-computed before creating the model).
	AvailableAgents []model.Agent
	// BuildPlanFn is the plan-builder function. Defaults to install.BuildPlan.
	// Tests inject a fake to avoid filesystem side effects.
	BuildPlanFn func(cat install.Catalog, intent install.Intent, opts install.Options) (install.Plan, error)
	// RunPlanFn executes the plan and sends progress/done messages to the
	// program. Tests inject a fake that immediately sends a doneMsg.
	RunPlanFn func(plan install.Plan, bridge *progressBridge, send func(tea.Msg))

	// Starters is the list of available starters (pre-computed from catalog.AllStarters()).
	Starters []model.Starter
	// BackupDir is the directory where backups are stored.
	BackupDir string
	// RunUninstall invokes the headless uninstall executor with the given flags
	// and writer. Returns 0 on success, 1 on failure.
	// Wraps RunHeadlessUninstall with real catalog/registry injected.
	RunUninstall func(flags headless.ParsedUninstallFlags, w io.Writer) int
	// RunStarter invokes the starter-add executor for the given starter ID.
	// projectPath is the target project directory; agents are the focal agents.
	// Returns 0 on success, 1 on failure.
	RunStarter func(starterID, projectPath string, agents []model.Agent, w io.Writer) int
}

// hubItem is one item in the welcome hub menu.
type hubItem struct {
	label string
}

// hubItems defines the canonical ordered list of hub menu options.
var hubItems = []hubItem{
	{label: "Install"},
	{label: "Starters"},
	{label: "Manage backups"},
	{label: "Uninstall"},
	{label: "Update stack"},
	{label: "Quit"},
}

// Hub item indices (single source of truth).
const (
	hubInstall  = 0
	hubStarters = 1
	hubBackups  = 2
	hubUninstall = 3
	hubUpdate   = 4
	hubQuit     = 5
)

// Model is the top-level Bubbletea model for the install flow.
type Model struct {
	Screen    Screen
	Cursor    int
	Selection Selection

	// Available* fields are resolved at construction from deps.
	AvailableAgents   []model.Agent
	AvailableHarnesses []model.Harness // populated on ScreenCustomPicker entry

	// Review state.
	ResolvedIDs []string // step IDs from BuildPlan, in topological order
	ReviewErr   error    // non-nil if BuildPlan failed

	// Progress state (ScreenInstalling).
	bridge   *progressBridge
	stepRows []stepRow // one per Apply step in the plan

	// Completion state (ScreenComplete).
	ExecutionResult pipeline.ExecutionResult

	// Hub navigation state.
	prevScreen Screen  // set when entering a child screen from the hub
	hubNotice  string  // inline message on hub (e.g. "coming soon")

	// Backups screen state.
	backups backupsState

	// Uninstall screen state.
	uninstallSel      uninstallSelections
	uninstallLines    []string
	uninstallDone     bool
	uninstallExitCode int
	uninstallBridge   *uninstallProgressBridge

	// Injected dependencies.
	deps ModelDeps
}

// stepRow holds the display state of one install step.
type stepRow struct {
	stepID string
	status pipeline.StepStatus
	err    error
}

// modeOrder is the single source of truth for the order in which install modes
// are presented on ScreenMode. ScreenMode (key handling and render) reads this
// slice so the order can never drift between the two.
var modeOrder = []model.InstallMode{model.ModeLite, model.ModeFull, model.ModeCustom}

// defaultModeCursor returns the index into modeOrder of the mode pre-selected
// when the user reaches ScreenMode. Full is the recommended baseline (sustrato
// + fundación guiada), so the radio starts on it. Computed from modeOrder so a
// reorder of the modes cannot leave a stale hardcoded index behind.
func defaultModeCursor() int {
	for i, mode := range modeOrder {
		if mode == model.ModeFull {
			return i
		}
	}
	return 0
}

// newModel creates a Model with the provided dependencies.
func newModel(deps ModelDeps) Model {
	if deps.BuildPlanFn == nil {
		deps.BuildPlanFn = install.BuildPlan
	}
	if deps.RunPlanFn == nil {
		deps.RunPlanFn = defaultRunPlan
	}
	return Model{
		Screen:          ScreenWelcome,
		AvailableAgents: deps.AvailableAgents,
		deps:            deps,
	}
}

// Init is called once when the program starts.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles all messages and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle uninstall-specific messages before the generic switch.
	if updated, cmd, handled := m.applyUninstallMsg(msg); handled {
		return updated, cmd
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		return m.handleKey(msg)

	case progressMsg:
		m.applyProgressEvent(msg.event)
		return m, listenCmd(m.bridge)

	case doneMsg:
		m.ExecutionResult = msg.result
		m.Screen = ScreenComplete
		return m, nil
	}

	return m, nil
}

// handleKey routes keyboard events to the active screen.
func (m Model) handleKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Normalización cross-platform de la barra espaciadora. En Windows el driver
	// de consola de bubbletea (key_windows.go: VK_SPACE → KeyRunes) entrega el
	// espacio como KeyRunes con rune ' ', mientras que en unix el parser ANSI lo
	// entrega como KeySpace. Sin esta normalización, en Windows el espacio cae en
	// el case de navegación (tea.KeyUp, tea.KeyRunes) de ScreenAgents, que lo
	// descarta, y el case tea.KeySpace queda inalcanzable: el toggle nunca
	// dispara (y de paso el Enter queda bloqueado por el guard de selección).
	// Unificamos a KeySpace para que cada pantalla maneje el espacio igual en
	// ambas plataformas.
	if key.Type == tea.KeyRunes && len(key.Runes) == 1 && key.Runes[0] == ' ' {
		key.Type = tea.KeySpace
		key.Runes = nil
	}

	switch m.Screen {

	case ScreenWelcome:
		return m.handleWelcomeKey(key)


	case ScreenDetection:
		switch key.Type {
		case tea.KeyEnter:
			m.Screen = ScreenAgents
			m.Cursor = 0
		case tea.KeyEsc:
			if s, ok := prevScreen(m.Screen); ok {
				m.Screen = s
			}
		}

	case ScreenAgents:
		switch key.Type {
		case tea.KeyUp, tea.KeyRunes:
			if key.Type == tea.KeyUp || string(key.Runes) == "k" {
				if m.Cursor > 0 {
					m.Cursor--
				}
			}
		case tea.KeyDown:
			if m.Cursor < len(m.AvailableAgents)-1 {
				m.Cursor++
			}
		case tea.KeySpace:
			if m.Cursor < len(m.AvailableAgents) {
				agent := m.AvailableAgents[m.Cursor]
				m.Selection.Agents = toggleAgent(m.Selection.Agents, agent)
			}
		case tea.KeyEnter:
			if len(m.Selection.Agents) == 0 {
				// Guard: at least one agent required.
				break
			}
			m.Screen = ScreenMode
			m.Cursor = defaultModeCursor() // Full por defecto
		case tea.KeyEsc:
			if s, ok := prevScreen(m.Screen); ok {
				m.Screen = s
				m.Cursor = 0
			}
		}

	case ScreenMode:
		modes := modeOrder
		switch key.Type {
		case tea.KeyUp:
			if m.Cursor > 0 {
				m.Cursor--
			}
		case tea.KeyDown:
			if m.Cursor < len(modes)-1 {
				m.Cursor++
			}
		case tea.KeyEnter:
			m.Selection.Mode = modes[m.Cursor]
			if m.Selection.Mode == model.ModeCustom {
				// Populate harness list from catalog.
				if m.deps.Catalog != nil {
					all := m.deps.Catalog.ForMode(model.ModeCustom)
					m.AvailableHarnesses = filterHarnessesByAgents(all, m.Selection.Agents)
					// C-21: permissions es security-first — no desactivable en
					// Custom. Lo pre-seleccionamos para que arranque marcado y el
					// usuario vea que se instala siempre.
					if isHarnessAvailable(m.AvailableHarnesses, install.SecurityFirstHarnessID) &&
						!isStringSelected(m.Selection.CustomHarnesses, install.SecurityFirstHarnessID) {
						m.Selection.CustomHarnesses = append(m.Selection.CustomHarnesses, install.SecurityFirstHarnessID)
					}
				}
				m.Screen = ScreenCustomPicker
				m.Cursor = 0
			} else {
				// Lite/Full: show ScreenPermissions if at least one tier-capable agent is selected.
				// D8: if no tier-capable agent, skip directly to review.
				if anyTierCapable(m.Selection.Agents) {
					return m.enterPermissions(), nil
				}
				return m.enterReview()
			}
		case tea.KeyEsc:
			if s, ok := prevScreen(m.Screen); ok {
				m.Screen = s
				m.Cursor = 0
			}
		}

	case ScreenCustomPicker:
		switch key.Type {
		case tea.KeyUp:
			if m.Cursor > 0 {
				m.Cursor--
			}
		case tea.KeyDown:
			if m.Cursor < len(m.AvailableHarnesses)-1 {
				m.Cursor++
			}
		case tea.KeySpace:
			if m.Cursor < len(m.AvailableHarnesses) {
				id := m.AvailableHarnesses[m.Cursor].ID
				// C-21: permissions es security-first — ignorar el toggle, no se
				// puede desmarcar.
				if id != install.SecurityFirstHarnessID {
					m.Selection.CustomHarnesses = toggleString(m.Selection.CustomHarnesses, id)
				}
			}
		case tea.KeyEnter:
			// D8: only show ScreenPermissions when at least one tier-capable agent is selected.
			if anyTierCapable(m.Selection.Agents) {
				return m.enterPermissions(), nil
			}
			return m.enterReview()
		case tea.KeyEsc:
			m.Screen = ScreenMode
			m.Cursor = defaultModeCursor() // Full por defecto al volver al modo
		}

	case ScreenPermissions:
		return m.handlePermissionsKey(key)

	case ScreenStarters:
		return m.handleStartersKey(key)

	case ScreenBackups:
		return m.handleBackupsKey(key)

	case ScreenUninstallAgents, ScreenUninstallMode, ScreenUninstallStrategy:
		return m.handleUninstallFlowKey(key)

	case ScreenUninstallConfirm:
		return m.handleUninstallConfirmKey(key)

	case ScreenUninstalling:
		return m.handleUninstallingKey(key)

	case ScreenReview:
		switch key.Type {
		case tea.KeyEnter:
			if m.ReviewErr != nil {
				break // stay on review until user fixes selection
			}
			return m.startInstall()
		case tea.KeyEsc:
			// Go back: if tier-capable agents were selected, go back to ScreenPermissions.
			// Otherwise go back to ScreenCustomPicker (Custom) or ScreenMode (Lite/Full).
			if anyTierCapable(m.Selection.Agents) {
				m.Screen = ScreenPermissions
				m.Cursor = defaultTierCursor()
			} else if m.Selection.Mode == model.ModeCustom {
				m.Screen = ScreenCustomPicker
				m.Cursor = 0
			} else {
				m.Screen = ScreenMode
				m.Cursor = defaultModeCursor()
			}
		}

	case ScreenComplete:
		switch {
		case key.Type == tea.KeyRunes && string(key.Runes) == "q":
			return m, tea.Quit
		case key.Type == tea.KeyCtrlC:
			return m, tea.Quit
		}
	}

	return m, nil
}

// enterReview calls BuildPlan and transitions to ScreenReview. Pure (no side effects).
func (m Model) enterReview() (tea.Model, tea.Cmd) {
	m.ResolvedIDs = nil
	m.ReviewErr = nil

	if m.deps.Catalog != nil && m.deps.Registry != nil {
		intent := m.Selection.BuildIntent()
		opts := install.Options{
			HomeDir:  m.deps.HomeDir,
			Registry: m.deps.Registry,
			Profile:  system.PlatformProfile{OS: runtime.GOOS},
		}
		plan, err := m.deps.BuildPlanFn(m.deps.Catalog, intent, opts)
		if err != nil {
			m.ReviewErr = err
		} else {
			for _, s := range plan.Apply {
				m.ResolvedIDs = append(m.ResolvedIDs, s.ID())
			}
		}
	}

	m.Screen = ScreenReview
	return m, nil
}

// startInstall launches the orchestrator goroutine and transitions to ScreenInstalling.
func (m Model) startInstall() (tea.Model, tea.Cmd) {
	m.bridge = newProgressBridge(64)

	// Build step rows from resolved IDs.
	m.stepRows = make([]stepRow, len(m.ResolvedIDs))
	for i, id := range m.ResolvedIDs {
		m.stepRows[i] = stepRow{stepID: id, status: pipeline.StepStatusPending}
	}

	m.Screen = ScreenInstalling

	// Build the real plan with progress wired.
	intent := m.Selection.BuildIntent()
	opts := install.Options{
		HomeDir:    m.deps.HomeDir,
		Registry:   m.deps.Registry,
		OnProgress: m.bridge.OnProgress,
		Profile:    system.PlatformProfile{OS: runtime.GOOS},
	}

	var plan install.Plan
	var buildErr error
	if m.deps.Catalog != nil && m.deps.Registry != nil {
		plan, buildErr = m.deps.BuildPlanFn(m.deps.Catalog, intent, opts)
	}

	if buildErr != nil {
		// This should not happen (we already validated in review), but handle it.
		result := pipeline.ExecutionResult{Err: buildErr}
		m.ExecutionResult = result
		m.Screen = ScreenComplete
		return m, nil
	}

	// ── Pre-flight dependency gate ────────────────────────────────────────────
	// Check that all runtimes required by the selected harnesses are present
	// before starting the orchestrator. On missing deps, abort to ScreenComplete
	// with an error — no filesystem writes, no rollback.
	if gateErr := m.checkPreflightDeps(); gateErr != nil {
		m.ExecutionResult = pipeline.ExecutionResult{Err: gateErr}
		m.Screen = ScreenComplete
		return m, nil
	}

	// Launch the runner goroutine.
	bridge := m.bridge // capture for goroutine
	sendFn := func(msg tea.Msg) {} // placeholder — real send set by Program
	_ = sendFn
	go m.deps.RunPlanFn(plan, bridge, nil)

	return m, listenCmd(m.bridge)
}

// applyProgressEvent updates the step row matching the event's StepID.
func (m *Model) applyProgressEvent(ev pipeline.ProgressEvent) {
	for i := range m.stepRows {
		if m.stepRows[i].stepID == ev.StepID {
			m.stepRows[i].status = ev.Status
			m.stepRows[i].err = ev.Err
			return
		}
	}
}

// View renders the current screen.
func (m Model) View() string {
	switch m.Screen {
	case ScreenWelcome:
		return m.viewWelcome()
	case ScreenDetection:
		return m.viewDetection()
	case ScreenAgents:
		return m.viewAgents()
	case ScreenMode:
		return m.viewMode()
	case ScreenCustomPicker:
		return m.viewCustomPicker()
	case ScreenPermissions:
		return m.viewPermissions()
	case ScreenReview:
		return m.viewReview()
	case ScreenInstalling:
		return m.viewInstalling()
	case ScreenComplete:
		return m.viewComplete()
	case ScreenStarters:
		return m.viewStarters()
	case ScreenBackups:
		return m.viewBackups()
	case ScreenUninstallAgents, ScreenUninstallMode, ScreenUninstallStrategy:
		return m.viewUninstallFlow()
	case ScreenUninstallConfirm:
		return m.viewUninstallConfirm()
	case ScreenUninstalling:
		return m.viewUninstalling()
	default:
		return ""
	}
}

// ── Views (minimal Lipgloss layout, no cosmetic theme) ────────────────────────

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// handleWelcomeKey handles keyboard input on the ScreenWelcome hub.
func (m Model) handleWelcomeKey(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.Type {
	case tea.KeyUp:
		if m.Cursor > 0 {
			m.Cursor--
		}
		m.hubNotice = ""
	case tea.KeyDown:
		if m.Cursor < len(hubItems)-1 {
			m.Cursor++
		}
		m.hubNotice = ""
	case tea.KeyEnter:
		m.hubNotice = ""
		switch m.Cursor {
		case hubInstall:
			// Enter existing install flow.
			m.prevScreen = ScreenWelcome
			m.Screen = ScreenDetection
		case hubStarters:
			m.prevScreen = ScreenWelcome
			return m.enterStarters()
		case hubBackups:
			m.prevScreen = ScreenWelcome
			return m.enterBackups()
		case hubUninstall:
			m.prevScreen = ScreenWelcome
			m.Screen = ScreenUninstallAgents
			m.Cursor = 0
		case hubUpdate:
			m.hubNotice = "coming soon"
		case hubQuit:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) viewWelcome() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("jr-stack") + "\n\n")
	for i, item := range hubItems {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}
		sb.WriteString(cursor + item.label + "\n")
	}
	if m.hubNotice != "" {
		sb.WriteString("\n" + dimStyle.Render(m.hubNotice) + "\n")
	}
	sb.WriteString("\nUp/Down = navigate  Enter = select\n")
	return sb.String()
}

func (m Model) viewDetection() string {
	return titleStyle.Render("Detecting environment") + "\n\n" +
		"Press Enter to continue.\n"
}

func (m Model) viewAgents() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Select agent(s)") + "\n\n")
	for i, a := range m.AvailableAgents {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}
		checked := "[ ]"
		if isAgentSelected(m.Selection.Agents, a) {
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

func (m Model) viewMode() string {
	modes := modeOrder
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Choose install mode") + "\n\n")
	for i, mode := range modes {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}
		sb.WriteString(fmt.Sprintf("%s%s\n", cursor, mode))
	}
	sb.WriteString("\nEnter = select  Esc = back\n")
	return sb.String()
}

func (m Model) viewCustomPicker() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Select harnesses") + "\n\n")
	for i, h := range m.AvailableHarnesses {
		cursor := "  "
		if m.Cursor == i {
			cursor = "> "
		}
		checked := "[ ]"
		if isStringSelected(m.Selection.CustomHarnesses, h.ID) {
			checked = selectedStyle.Render("[x]")
		}
		// C-21: permissions es security-first — se muestra forzado ([x] fijo)
		// con un sufijo que explica por qué no se puede desmarcar.
		suffix := ""
		if h.ID == install.SecurityFirstHarnessID {
			checked = selectedStyle.Render("[x]")
			suffix = dimStyle.Render(" (requerido — security-first)")
		}
		sb.WriteString(fmt.Sprintf("%s%s %s%s\n", cursor, checked, h.Name, suffix))
	}
	sb.WriteString("\nSpace = toggle  Enter = continue  Esc = back\n")
	return sb.String()
}

func (m Model) viewReview() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Review") + "\n\n")
	if m.ReviewErr != nil {
		sb.WriteString(errorStyle.Render("Error: "+m.ReviewErr.Error()) + "\n\n")
		sb.WriteString("Esc = go back\n")
		return sb.String()
	}
	sb.WriteString("Steps to install (topological order):\n")
	for _, id := range m.ResolvedIDs {
		sb.WriteString("  • " + id + "\n")
	}
	if len(m.ResolvedIDs) == 0 {
		sb.WriteString(dimStyle.Render("  (nothing to install)\n"))
	}
	sb.WriteString("\nEnter = install  Esc = back\n")
	return sb.String()
}

func (m Model) viewInstalling() string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Installing") + "\n\n")
	for _, row := range m.stepRows {
		icon := statusIcon(row.status)
		sb.WriteString(fmt.Sprintf("  %s %s\n", icon, row.stepID))
	}
	sb.WriteString("\n")
	return sb.String()
}

func (m Model) viewComplete() string {
	var sb strings.Builder
	if m.ExecutionResult.Err != nil {
		sb.WriteString(titleStyle.Render("Installation failed") + "\n\n")
		sb.WriteString(errorStyle.Render(m.ExecutionResult.Err.Error()) + "\n\n")
		if len(m.ExecutionResult.Rollback.Steps) > 0 {
			sb.WriteString("Rollback: ")
			if m.ExecutionResult.Rollback.Success {
				sb.WriteString(selectedStyle.Render("succeeded") + "\n")
			} else {
				sb.WriteString(errorStyle.Render("failed") + "\n")
			}
		}
	} else {
		sb.WriteString(titleStyle.Render("Installation complete!") + "\n\n")
		// C-32: collect degraded rows for an honest summary.
		var degraded []stepRow
		for _, row := range m.stepRows {
			if row.status == pipeline.StepStatusDegraded {
				degraded = append(degraded, row)
			}
		}
		if len(degraded) == 0 {
			sb.WriteString(selectedStyle.Render("All steps succeeded.") + "\n\n")
		} else {
			sb.WriteString(selectedStyle.Render("Installation succeeded with degraded harnesses.") + "\n\n")
			sb.WriteString(dimStyle.Render("Degraded (best-effort) harnesses:") + "\n")
			for _, row := range degraded {
				reason := ""
				if row.err != nil {
					reason = ": " + row.err.Error()
				}
				sb.WriteString(fmt.Sprintf("  %s %s%s\n", dimStyle.Render("⚠"), row.stepID, reason))
			}
			sb.WriteString("\n")
		}
	}
	sb.WriteString("Press q to quit.\n")
	return sb.String()
}

func statusIcon(s pipeline.StepStatus) string {
	switch s {
	case pipeline.StepStatusRunning:
		return "⟳"
	case pipeline.StepStatusSucceeded:
		return selectedStyle.Render("✓")
	case pipeline.StepStatusFailed:
		return errorStyle.Render("✗")
	case pipeline.StepStatusDegraded:
		// C-32: distinct warning glyph for best-effort degraded steps.
		return dimStyle.Render("⚠")
	case pipeline.StepStatusRolledBack:
		return dimStyle.Render("↩")
	default:
		return "·"
	}
}

// ── Selection helpers ─────────────────────────────────────────────────────────

// The security-first harness ID lives in the install package as
// install.SecurityFirstHarnessID — the single source of truth for "what is
// forced" (C-24). The TUI references it directly instead of keeping a local copy.

// isHarnessAvailable reports whether a harness with the given ID is present in
// the (already agent-filtered) available list.
func isHarnessAvailable(harnesses []model.Harness, id string) bool {
	for _, h := range harnesses {
		if h.ID == id {
			return true
		}
	}
	return false
}

func toggleAgent(agents []model.Agent, a model.Agent) []model.Agent {
	for i, existing := range agents {
		if existing == a {
			return append(agents[:i:i], agents[i+1:]...)
		}
	}
	return append(agents, a)
}

func toggleString(strs []string, s string) []string {
	for i, existing := range strs {
		if existing == s {
			return append(strs[:i:i], strs[i+1:]...)
		}
	}
	return append(strs, s)
}

func isAgentSelected(agents []model.Agent, a model.Agent) bool {
	for _, existing := range agents {
		if existing == a {
			return true
		}
	}
	return false
}

func isStringSelected(strs []string, s string) bool {
	for _, existing := range strs {
		if existing == s {
			return true
		}
	}
	return false
}
