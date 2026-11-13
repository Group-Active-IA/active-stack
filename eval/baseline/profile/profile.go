// Package profile define los perfiles de medicion por cliente e interfaz
// (tarea 1.2 de opsx-baseline): el perfil equivalente para edicion y la serie
// separada de Estricto. Son datos mas validacion; no ejecutan ningun cliente.
// La verificacion con el cliente real en Windows y Linux queda pendiente
// (ver Pending).
//
// Convenciones de manifest.Permissions para estos perfiles:
//
//	Claude, terminal: Mode = valor de --permission-mode; Flags incluye
//	  "--permission-mode <modo>" y un "--allowedTools <Herramienta(patron)>" por
//	  permiso de allow.
//	Claude, IDE (VS Code): sin flags de CLI; Mode = initialPermissionMode y las
//	  reglas de allow viven en el settings.json compartido con la CLI, escritas
//	  "settings.json:permissions.allow=<Herramienta(patron)>".
//	Codex: Mode = "<sandbox_mode>/<approval_policy>". Terminal usa
//	  "--sandbox X" y "--ask-for-approval Y"; el IDE comparte config.toml con la
//	  CLI y se escribe "config.toml:sandbox_mode=X" y "config.toml:approval_policy=Y".
package profile

import (
	"fmt"
	"strings"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
)

// Kind distingue el perfil equivalente para edicion de la serie Estricto.
type Kind string

// Tipos de perfil.
const (
	KindEquivalent Kind = "equivalente_para_edicion"
	KindStrict     Kind = "estricto"
)

// Source es una afirmacion de la documentacion oficial que respalda el perfil.
type Source struct {
	URL       string
	Consulted string
	Quote     string
}

// Profile es un perfil de medicion. Permissions es lo que el manifiesto debe
// declarar para las celdas que lo usan.
type Profile struct {
	ID          string
	Client      manifest.Client
	Interface   manifest.Interface
	Kind        Kind
	Permissions manifest.Permissions
	Sources     []Source
	// Limits son limites documentados del perfil (que NO restringe).
	Limits []string
	// Unverified son afirmaciones que la documentacion no confirma.
	Unverified []string
}

const consulted = "2026-11-11"

var (
	claudeModesSource = Source{
		URL:       "https://code.claude.com/docs/en/permission-modes",
		Consulted: consulted,
		Quote:     "`acceptEdits` mode lets Claude create and edit files in your working directory without prompting.",
	}
	claudeRulesSource = Source{
		URL:       "https://code.claude.com/docs/en/permissions",
		Consulted: consulted,
		Quote:     "`path` or `./path`: Path relative to current directory (`Read(*.env)` -> `<cwd>/*.env`)",
	}
	claudeHeadlessSource = Source{
		URL:       "https://code.claude.com/docs/en/headless",
		Consulted: consulted,
		Quote:     "In a Manual mode `claude -p` run, a call that needs approval is denied unless something approves it",
	}
	claudeVSCodeSource = Source{
		URL:       "https://code.claude.com/docs/en/vs-code",
		Consulted: consulted,
		Quote:     "`initialPermissionMode`: Controls approval prompts for new conversations: `default`, `plan`, `acceptEdits`, or `bypassPermissions`. `manual` is an alias for `default`",
	}
	claudeVSCodeSharedSource = Source{
		URL:       "https://code.claude.com/docs/en/vs-code",
		Consulted: consulted,
		Quote:     "Claude Code settings in `~/.claude/settings.json`: shared between the extension and CLI.",
	}
	codexApprovalsSource = Source{
		URL:       "https://learn.chatgpt.com/docs/agent-approvals-security",
		Consulted: consulted,
		Quote:     "Auto (preset): `--sandbox workspace-write --ask-for-approval on-request`: Edits and commands run in the workspace. Edits outside it or network use need approval.",
	}
	codexStrictSource = Source{
		URL:       "https://learn.chatgpt.com/docs/agent-approvals-security",
		Consulted: consulted,
		Quote:     "Safe read-only browsing: `--sandbox read-only --ask-for-approval on-request`: Reads and commands stay in the read-only sandbox. Actions outside it can need approval.",
	}
	codexExecSource = Source{
		URL:       "https://learn.chatgpt.com/docs/non-interactive-mode",
		Consulted: consulted,
		Quote:     "By default, `codex exec` runs in a read-only sandbox. `--sandbox workspace-write` allows edits.",
	}
	codexIDESharedSource = Source{
		URL:       "https://learn.chatgpt.com/docs/ide/settings",
		Consulted: consulted,
		Quote:     "agent settings are \"shared with Codex CLI\"; sandbox and permissions are listed among the shared Codex settings",
	}
)

var claudeCommonLimits = []string{
	"acceptEdits tambien auto-aprueba mkdir, touch, rm, rmdir, mv, cp y sed dentro del directorio de trabajo",
	"las escrituras a rutas protegidas (.git, .claude, .vscode, .idea, ...) siguen pidiendo aprobacion y las reglas allow no las pre-aprueban",
	"--bare no se usa: omitiria CLAUDE.md, skills, hooks y subagentes que el flujo medido necesita",
}

var codexCommonLimits = []string{
	"workspace-write incluye los directorios temporales (por ejemplo /tmp) ademas del directorio de trabajo, asi que no escribe solo dentro del checkout",
	"las rutas protegidas .git, .codex, .agents y .aws quedan de solo lectura de forma recursiva",
	"el acceso a red esta apagado por defecto en workspace-write",
}

func claudeProfile(i manifest.Interface, k Kind) Profile {
	p := Profile{Client: manifest.ClientClaude, Interface: i, Kind: k}
	mode := "default"
	if k == KindEquivalent {
		mode = "acceptEdits"
	}
	p.ID = fmt.Sprintf("claude-%s-%s", i, k)
	p.Limits = append([]string{}, claudeCommonLimits...)
	switch i {
	case manifest.InterfaceTerminal:
		flags := []string{"--permission-mode " + mode}
		if k == KindEquivalent {
			flags = append(flags, "--allowedTools Edit(./**)", "--allowedTools Read(./**)", "--allowedTools Bash(go test *)")
		}
		p.Permissions = manifest.Permissions{Mode: mode, Flags: flags}
		p.Sources = []Source{claudeModesSource, claudeRulesSource, claudeHeadlessSource}
		p.Unverified = []string{"las reglas --allowedTools Edit(./**) se anclan al directorio de trabajo de la sesion; la ancla exacta con el cliente real no esta verificada"}
	case manifest.InterfaceIDE:
		flags := []string{"vscode:initialPermissionMode=" + mode}
		if k == KindEquivalent {
			flags = append(flags, "settings.json:permissions.allow=Edit(./**)", "settings.json:permissions.allow=Read(./**)", "settings.json:permissions.allow=Bash(go test *)")
		}
		p.Permissions = manifest.Permissions{Mode: mode, Flags: flags}
		p.Sources = []Source{claudeModesSource, claudeVSCodeSource, claudeVSCodeSharedSource}
		p.Unverified = []string{
			"el prefijo exacto de la clave en settings.json de VS Code (la pagina lista `initialPermissionMode`)",
			"la documentacion dice que VS Code lee initialPermissionMode de los ajustes de usuario y ignora los de workspace",
		}
	}
	if k == KindStrict {
		p.Sources = append(p.Sources, claudeHeadlessSource)
		p.Unverified = append(p.Unverified, "el contrato de Estricto de 01 se leyo solo para Codex; para Claude se define como Manual (default) sin reglas allow")
	}
	return p
}

func codexProfile(i manifest.Interface, k Kind) Profile {
	p := Profile{Client: manifest.ClientCodex, Interface: i, Kind: k}
	sandbox := "read-only"
	if k == KindEquivalent {
		sandbox = "workspace-write"
	}
	const approval = "on-request"
	p.ID = fmt.Sprintf("codex-%s-%s", i, k)
	p.Limits = append([]string{}, codexCommonLimits...)
	mode := sandbox + "/" + approval
	switch i {
	case manifest.InterfaceTerminal:
		p.Permissions = manifest.Permissions{Mode: mode, Flags: []string{"--sandbox " + sandbox, "--ask-for-approval " + approval}}
		p.Sources = []Source{codexApprovalsSource, codexExecSource}
		p.Unverified = []string{"la pagina de `codex exec` no documenta --ask-for-approval; en exec la politica puede fijarse con approval_policy en config.toml"}
	case manifest.InterfaceIDE:
		p.Permissions = manifest.Permissions{Mode: mode, Flags: []string{"config.toml:sandbox_mode=" + sandbox, "config.toml:approval_policy=" + approval}}
		p.Sources = []Source{codexApprovalsSource, codexIDESharedSource}
		p.Unverified = []string{
			"la correspondencia entre los modos con nombre del IDE (Chat, Agent, Agent (Full Access)) y sandbox_mode/approval_policy no esta en la documentacion oficial consultada",
			"el path de config.toml compartido no figura en la pagina de ajustes del IDE",
		}
	}
	if k == KindStrict {
		p.Sources = append(p.Sources, codexStrictSource)
	}
	return p
}

// Profiles devuelve los perfiles definidos.
func Profiles() []Profile {
	var out []Profile
	for _, i := range []manifest.Interface{manifest.InterfaceTerminal, manifest.InterfaceIDE} {
		for _, k := range []Kind{KindEquivalent, KindStrict} {
			out = append(out, claudeProfile(i, k), codexProfile(i, k))
		}
	}
	return out
}

// Lookup busca un perfil.
func Lookup(c manifest.Client, i manifest.Interface, k Kind) (Profile, bool) {
	for _, p := range Profiles() {
		if p.Client == c && p.Interface == i && p.Kind == k {
			return p, true
		}
	}
	return Profile{}, false
}

// claudeAllow son las reglas de allow de un perfil de Claude.
type claudeAllow struct{ tool, spec string }

const (
	claudeModeFlagTerminal = "--permission-mode "
	claudeModeFlagIDE      = "vscode:initialPermissionMode="
	claudeAllowTerminal    = "--allowedTools "
	claudeAllowIDE         = "settings.json:permissions.allow="
)

// parseClaude interpreta los flags de un perfil de Claude. Solo admite el
// flag de modo y reglas allow; cualquier otro flag es un error.
func parseClaude(p Profile) (mode string, allow []claudeAllow, err error) {
	modePrefix, allowPrefix := claudeModeFlagTerminal, claudeAllowTerminal
	if p.Interface == manifest.InterfaceIDE {
		modePrefix, allowPrefix = claudeModeFlagIDE, claudeAllowIDE
	}
	modes := 0
	for _, f := range p.Permissions.Flags {
		switch {
		case strings.HasPrefix(f, modePrefix):
			modes++
			mode = strings.TrimPrefix(f, modePrefix)
		case strings.HasPrefix(f, allowPrefix):
			rule := strings.TrimPrefix(f, allowPrefix)
			open := strings.Index(rule, "(")
			if open <= 0 || !strings.HasSuffix(rule, ")") {
				return "", nil, fmt.Errorf("regla allow sin alcance: %q (Herramienta(patron) obligatorio)", rule)
			}
			allow = append(allow, claudeAllow{tool: rule[:open], spec: rule[open+1 : len(rule)-1]})
		default:
			return "", nil, fmt.Errorf("flag no admitido en un perfil de medicion de Claude: %q", f)
		}
	}
	if modes != 1 {
		return "", nil, fmt.Errorf("debe haber exactamente un flag de modo (%q), hay %d", modePrefix, modes)
	}
	return mode, allow, nil
}

// codexModes interpreta "<sandbox>/<approval>".
func codexModes(mode string) (sandbox, approval string, err error) {
	parts := strings.Split(mode, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("modo de Codex %q: se esperaba <sandbox_mode>/<approval_policy>", mode)
	}
	sandbox, approval = parts[0], parts[1]
	switch sandbox {
	case "read-only", "workspace-write":
	default:
		return "", "", fmt.Errorf("sandbox_mode %q no admitido para medir (solo read-only y workspace-write)", sandbox)
	}
	switch approval {
	case "on-request", "never":
	default:
		return "", "", fmt.Errorf("approval_policy %q no admitida (solo on-request y never; untrusted esta retirada)", approval)
	}
	return sandbox, approval, nil
}

// parseCodex interpreta los flags de un perfil de Codex y devuelve el sandbox
// y la politica que fijan.
func parseCodex(p Profile) (sandbox, approval string, err error) {
	sandboxPrefix, approvalPrefix := "--sandbox ", "--ask-for-approval "
	if p.Interface == manifest.InterfaceIDE {
		sandboxPrefix, approvalPrefix = "config.toml:sandbox_mode=", "config.toml:approval_policy="
	}
	var sb, ap int
	for _, f := range p.Permissions.Flags {
		switch {
		case strings.HasPrefix(f, sandboxPrefix):
			sb++
			sandbox = strings.TrimPrefix(f, sandboxPrefix)
		case strings.HasPrefix(f, approvalPrefix):
			ap++
			approval = strings.TrimPrefix(f, approvalPrefix)
		default:
			return "", "", fmt.Errorf("flag no admitido en un perfil de medicion de Codex: %q", f)
		}
	}
	if sb != 1 || ap != 1 {
		return "", "", fmt.Errorf("debe haber un flag de sandbox y uno de aprobacion, hay %d y %d", sb, ap)
	}
	return sandbox, approval, nil
}

func (p Profile) validateClaude() error {
	mode, allow, err := parseClaude(p)
	if err != nil {
		return err
	}
	if mode != p.Permissions.Mode {
		return fmt.Errorf("el flag de modo (%s) no coincide con el modo declarado (%s)", mode, p.Permissions.Mode)
	}
	switch mode {
	case "acceptEdits", "default":
	default:
		return fmt.Errorf("modo %q no admitido para medir (solo acceptEdits y default)", mode)
	}
	for _, a := range allow {
		switch a.tool {
		case "Edit", "Read":
			if !strings.HasPrefix(a.spec, "./") {
				return fmt.Errorf("%s(%s): el alcance debe ser relativo al checkout (./...)", a.tool, a.spec)
			}
		case "Bash":
			if strings.TrimSpace(a.spec) == "" || strings.TrimSpace(a.spec) == "*" {
				return fmt.Errorf("Bash(%s): un allow de shell sin comando acotado", a.spec)
			}
		default:
			return fmt.Errorf("herramienta %q no admitida en allow", a.tool)
		}
	}
	if p.Kind == KindStrict && len(allow) > 0 {
		return fmt.Errorf("el perfil Estricto no admite reglas allow")
	}
	return nil
}

func (p Profile) validateCodex() error {
	sandbox, approval, err := parseCodex(p)
	if err != nil {
		return err
	}
	modeSandbox, modeApproval, err := codexModes(p.Permissions.Mode)
	if err != nil {
		return err
	}
	if sandbox != modeSandbox || approval != modeApproval {
		return fmt.Errorf("los flags (%s/%s) no coinciden con el modo declarado (%s)", sandbox, approval, p.Permissions.Mode)
	}
	return nil
}

// Validate comprueba que el perfil sea coherente y no abra mas permisos que el
// pedido (bypass, acceso total, allow sin alcance, directorios extra).
func (p Profile) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("perfil sin id")
	}
	if p.Kind != KindEquivalent && p.Kind != KindStrict {
		return fmt.Errorf("%s: tipo %q desconocido", p.ID, p.Kind)
	}
	var err error
	switch p.Client {
	case manifest.ClientClaude:
		err = p.validateClaude()
	case manifest.ClientCodex:
		err = p.validateCodex()
	default:
		err = fmt.Errorf("cliente %q fuera de la matriz", p.Client)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", p.ID, err)
	}
	needs, err := p.EditInsideCheckoutNeedsApproval()
	if err != nil {
		return fmt.Errorf("%s: %w", p.ID, err)
	}
	if want := p.Kind == KindStrict; needs != want {
		return fmt.Errorf("%s: el tipo %s exige que una edicion dentro del checkout pida aprobacion = %v, pero el perfil da %v", p.ID, p.Kind, want, needs)
	}
	return nil
}

// EditInsideCheckoutNeedsApproval dice, segun la documentacion, si una edicion
// dentro del checkout pide aprobacion con este perfil.
func (p Profile) EditInsideCheckoutNeedsApproval() (bool, error) {
	switch p.Client {
	case manifest.ClientClaude:
		mode, allow, err := parseClaude(p)
		if err != nil {
			return false, err
		}
		switch mode {
		case "acceptEdits":
			return false, nil
		case "default":
			for _, a := range allow {
				if a.tool == "Edit" && strings.HasPrefix(a.spec, "./") {
					return false, nil
				}
			}
			return true, nil
		}
		return false, fmt.Errorf("modo %q fuera de los documentados para medir", mode)
	case manifest.ClientCodex:
		sandbox, _, err := codexModes(p.Permissions.Mode)
		if err != nil {
			return false, err
		}
		// workspace-write: las ediciones del workspace corren sin aprobacion con
		// on-request y con never; read-only no las permite sin aprobacion.
		return sandbox == "read-only", nil
	}
	return false, fmt.Errorf("cliente %q fuera de la matriz", p.Client)
}

// Compare contrasta los permisos efectivos observados con el perfil usando la
// comprobacion previa de manifest (modo, flags y opt-ins de contenido).
func Compare(p Profile, obs manifest.Observed) manifest.CellCheck {
	return manifest.CheckCell(manifest.Cell{
		Client: p.Client, Version: "0", MinVersion: "0",
		OS: manifest.OSLinux, Interface: p.Interface, Modality: manifest.ModalityAPI,
		Permissions: p.Permissions, Observed: obs,
	})
}

// CellMatches devuelve los motivos por los que la celda declarada no coincide
// con el perfil (cliente, interfaz o permisos).
func CellMatches(p Profile, c manifest.Cell) []string {
	var r []string
	if c.Client != p.Client {
		r = append(r, fmt.Sprintf("cliente %s distinto del perfil (%s)", c.Client, p.Client))
	}
	if c.Interface != p.Interface {
		r = append(r, fmt.Sprintf("interfaz %s distinta del perfil (%s)", c.Interface, p.Interface))
	}
	if c.Permissions.Mode != p.Permissions.Mode {
		r = append(r, fmt.Sprintf("modo %s distinto del perfil (%s)", c.Permissions.Mode, p.Permissions.Mode))
	}
	if !sameFlags(c.Permissions.Flags, p.Permissions.Flags) {
		r = append(r, fmt.Sprintf("flags %v distintos de los del perfil %v", c.Permissions.Flags, p.Permissions.Flags))
	}
	return r
}

func sameFlags(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	n := map[string]int{}
	for _, s := range a {
		n[s]++
	}
	for _, s := range b {
		n[s]--
	}
	for _, v := range n {
		if v != 0 {
			return false
		}
	}
	return true
}

// PendingItem es una verificacion real que sigue pendiente.
type PendingItem struct {
	ProfileID string
	OS        manifest.OS
	Check     string
}

// Pending lista las verificaciones con el cliente real que siguen pendientes:
// por perfil y sistema, que el modo y los flags efectivos coincidan con el
// manifiesto y, en el equivalente, que una edicion dentro del checkout no pida
// aprobacion (en Estricto, que la pida o quede bloqueada).
func Pending() []PendingItem {
	var out []PendingItem
	for _, p := range Profiles() {
		for _, os := range []manifest.OS{manifest.OSWindows, manifest.OSLinux} {
			check := fmt.Sprintf("iniciar %s real (%s) en %s, leer el modo y los flags efectivos y compararlos con el manifiesto; ", p.Client, p.Interface, os)
			if p.Kind == KindEquivalent {
				check += "comprobar que una edicion dentro del checkout no pide aprobacion"
			} else {
				check += "comprobar que una edicion dentro del checkout pide aprobacion o queda bloqueada y se registra"
			}
			out = append(out, PendingItem{ProfileID: p.ID, OS: os, Check: check})
		}
	}
	return out
}
