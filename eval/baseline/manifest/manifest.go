// Package manifest define el manifiesto versionado y la matriz de cohortes de
// la medicion de referencia (change opsx-baseline, tarea 1.1).
//
// El manifiesto es la unica fuente de version, interfaz, permisos y modalidad
// para los changes que reutilizan el protocolo (decision D10).
package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

// Client identifica el cliente de IA medido.
type Client string

// OS es el sistema operativo de la celda.
type OS string

// Interface es la interfaz REAL desde la que se ejecuta el cliente.
type Interface string

// Modality es la modalidad de acceso.
type Modality string

const (
	ClientClaude Client = "claude"
	ClientCodex  Client = "codex"

	OSWindows OS = "windows"
	OSLinux   OS = "linux"

	InterfaceTerminal Interface = "terminal"
	InterfaceIDE      Interface = "ide"

	ModalityAPI          Modality = "api"
	ModalitySubscription Modality = "subscription"
)

// Permissions son los permisos efectivos: modo y flags.
type Permissions struct {
	Mode  string   `json:"mode"`
	Flags []string `json:"flags"`
}

// Observed es lo comprobado con el cliente real en la comprobacion previa.
type Observed struct {
	Permissions Permissions `json:"permissions"`
	// Settings son las variables de entorno u opciones de configuracion
	// observadas (nombre -> valor). Sirve para detectar opt-ins de contenido.
	Settings map[string]string `json:"settings"`
	// CLIInIDETerminal indica que la CLI corre en la terminal integrada de un IDE.
	CLIInIDETerminal bool `json:"cli_in_ide_terminal"`
}

// Cell es la configuracion de una celda de la matriz.
type Cell struct {
	Client      Client            `json:"client"`
	Version     string            `json:"version"`
	MinVersion  string            `json:"min_version"`
	OS          OS                `json:"os"`
	Interface   Interface         `json:"interface"`
	Modality    Modality          `json:"modality"`
	Models      map[string]string `json:"models"` // rol -> modelo efectivo
	Effort      string            `json:"effort"`
	Tools       []string          `json:"tools"`
	Tier        string            `json:"tier"`
	Permissions Permissions       `json:"permissions"` // declarados
	Observed    Observed          `json:"observed"`
}

// CellStatus es el resultado de la validacion previa de una celda.
type CellStatus string

const (
	CellExecutable   CellStatus = "ejecutable"
	CellIncompatible CellStatus = "incompatible"
)

// CellCheck es el resultado de CheckCell.
type CellCheck struct {
	Status  CellStatus
	Reasons []string
	// Interface es la interfaz real de la cohorte en la que se agrega la celda.
	Interface Interface
}

// CheckCell valida una celda antes de iniciar corridas. Devuelve ejecutable o
// incompatible con todos los motivos detectados, en orden determinista.
func CheckCell(c Cell) CellCheck {
	reasons := checkDimensions(c)
	if r := checkVersion(c.Version, c.MinVersion); r != "" {
		reasons = append(reasons, r)
	}
	if strings.TrimSpace(c.Permissions.Mode) == "" {
		reasons = append(reasons, "permisos declarados vacíos: falta el modo de permisos")
	} else if r := checkPermissions(c.Permissions, c.Observed.Permissions); r != "" {
		reasons = append(reasons, r)
	}
	reasons = append(reasons, checkContentOptIns(c.Observed.Settings)...)
	if len(reasons) > 0 {
		return CellCheck{Status: CellIncompatible, Reasons: reasons}
	}
	return CellCheck{Status: CellExecutable, Interface: realInterface(c)}
}

func checkVersion(version, min string) string {
	v, ok := parseVersion(version)
	if !ok {
		return "versión no interpretable: " + version
	}
	m, ok := parseVersion(min)
	if !ok {
		return "versión mínima no interpretable: " + min
	}
	if compareVersions(v, m) < 0 {
		return "versión " + version + " menor que la mínima declarada " + min
	}
	return ""
}

// parseVersion interpreta "X.Y.Z" (cualquier cantidad de componentes numericos).
func parseVersion(s string) ([]int, bool) {
	if s == "" {
		return nil, false
	}
	parts := strings.Split(s, ".")
	out := make([]int, len(parts))
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil, false
		}
		out[i] = n
	}
	return out, true
}

func compareVersions(a, b []int) int {
	for i := 0; i < len(a) || i < len(b); i++ {
		var x, y int
		if i < len(a) {
			x = a[i]
		}
		if i < len(b) {
			y = b[i]
		}
		if x != y {
			if x < y {
				return -1
			}
			return 1
		}
	}
	return 0
}

func checkPermissions(declared, observed Permissions) string {
	if declared.Mode != observed.Mode {
		return "permisos efectivos distintos de los declarados: modo " + observed.Mode + " en lugar de " + declared.Mode
	}
	if !sameSet(declared.Flags, observed.Flags) {
		return "permisos efectivos distintos de los declarados: flags observados " + strings.Join(observed.Flags, " ") + " en lugar de " + strings.Join(declared.Flags, " ")
	}
	return ""
}

// sameSet compara multiconjuntos de strings sin importar el orden.
func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	count := map[string]int{}
	for _, s := range a {
		count[s]++
	}
	for _, s := range b {
		count[s]--
	}
	for _, n := range count {
		if n != 0 {
			return false
		}
	}
	return true
}

// TaskWeight es una tarea con su peso dentro de una cohorte.
type TaskWeight struct {
	TaskID string  `json:"task_id"`
	Weight float64 `json:"weight"`
}

// Cohort es una entrada de la matriz: una celda y las tareas que se miden en ella.
type Cohort struct {
	ID    string       `json:"id"`
	Cell  Cell         `json:"cell"`
	Tasks []TaskWeight `json:"tasks"`
}

// Prompt identifica el prompt congelado.
type Prompt struct {
	ID          string `json:"id"`
	Fingerprint string `json:"fingerprint"`
}

// Profile identifica el perfil de medicion; la huella no incluye secretos.
type Profile struct {
	Name              string `json:"name"`
	ConfigFingerprint string `json:"config_fingerprint"`
}

// Limits son los limites de la medicion. Los valores concretos dependen de
// la decision D3 (TBD); cero significa no fijado.
type Limits struct {
	MaxAttempts    int `json:"max_attempts"`
	TimeoutSeconds int `json:"timeout_seconds"`
}

// Manifest es el manifiesto versionado de una medicion.
type Manifest struct {
	ProtocolVersion     string `json:"protocol_version"`
	RubricVersion       string `json:"rubric_version"`
	FixtureRevision     string `json:"fixture_revision"`
	Prompt              Prompt `json:"prompt"`
	ActiveStackRevision string `json:"active_stack_revision"`
	// Variant identifica la variante medida: "actual", "modular" u otra.
	Variant string   `json:"variant"`
	Profile Profile  `json:"profile"`
	Seed    int64    `json:"seed"`
	Limits  Limits   `json:"limits"`
	Cohorts []Cohort `json:"cohorts"`
}

// Incompatible es una celda excluida con sus motivos.
type Incompatible struct {
	CohortID string
	Reasons  []string
}

// Planned es una cohorte ejecutable: Interface es la interfaz real en la que se
// agrega y DeclaredInterface la declarada en el manifiesto (se preservan ambas).
type Planned struct {
	CohortID          string
	Interface         Interface
	DeclaredInterface Interface
}

// Plan es el resultado de la comprobacion previa de toda la matriz.
type Plan struct {
	// Executable son las cohortes que entran al denominador.
	Executable []Planned
	// Incompatible son las cohortes fuera del denominador, listadas con su motivo.
	Incompatible []Incompatible
}

// Plan evalua cada cohorte antes de iniciar corridas. Valida el manifiesto primero.
func (m Manifest) Plan() (Plan, error) {
	if err := m.Validate(); err != nil {
		return Plan{}, err
	}
	var p Plan
	for _, c := range m.Cohorts {
		check := CheckCell(c.Cell)
		if check.Status == CellExecutable {
			p.Executable = append(p.Executable, Planned{CohortID: c.ID, Interface: check.Interface, DeclaredInterface: c.Cell.Interface})
			continue
		}
		p.Incompatible = append(p.Incompatible, Incompatible{CohortID: c.ID, Reasons: check.Reasons})
	}
	return p, nil
}

// RunStatus es el estado de una corrida iniciada. Una celda incompatible no
// tiene corridas y por eso no figura aqui (ver CellIncompatible).
type RunStatus string

// CountsInCost, el resumen y su calculo pertenecen a la tarea 3.1.
const (
	RunAccepted           RunStatus = "aceptada"
	RunRejectedQuality    RunStatus = "rechazada_por_calidad"
	RunBlockedPermissions RunStatus = "bloqueada_por_permisos"
	RunTimeout            RunStatus = "timeout"
	RunClientFailure      RunStatus = "fallo_del_cliente"
)

// Valid indica si el estado es uno de los definidos.
func (s RunStatus) Valid() bool {
	switch s {
	case RunAccepted, RunRejectedQuality, RunBlockedPermissions, RunTimeout, RunClientFailure:
		return true
	}
	return false
}

// Validate comprueba la estructura del manifiesto (no la compatibilidad de
// las celdas, que resuelve Plan).
func (m Manifest) Validate() error {
	if m.ProtocolVersion == "" {
		return fmt.Errorf("manifiesto: falta protocol_version")
	}
	if m.RubricVersion == "" {
		return fmt.Errorf("manifiesto: falta rubric_version")
	}
	if m.FixtureRevision == "" {
		return fmt.Errorf("manifiesto: falta fixture_revision")
	}
	if strings.TrimSpace(m.Variant) == "" {
		return fmt.Errorf("manifiesto: falta variant (actual, modular u otra)")
	}
	if len(m.Cohorts) == 0 {
		return fmt.Errorf("manifiesto: no hay cohortes")
	}
	seen := map[string]bool{}
	for _, c := range m.Cohorts {
		if c.ID == "" {
			return fmt.Errorf("manifiesto: cohorte sin id")
		}
		if seen[c.ID] {
			return fmt.Errorf("manifiesto: id de cohorte duplicado %q", c.ID)
		}
		seen[c.ID] = true
		if len(c.Tasks) == 0 {
			return fmt.Errorf("manifiesto: la cohorte %q no tiene tareas", c.ID)
		}
		taskIDs := map[string]bool{}
		for _, t := range c.Tasks {
			if t.TaskID == "" {
				return fmt.Errorf("manifiesto: la cohorte %q tiene una tarea sin task_id", c.ID)
			}
			if taskIDs[t.TaskID] {
				return fmt.Errorf("manifiesto: task_id duplicado %q en la cohorte %q", t.TaskID, c.ID)
			}
			taskIDs[t.TaskID] = true
			if t.Weight <= 0 {
				return fmt.Errorf("manifiesto: la tarea %q de la cohorte %q tiene peso no positivo", t.TaskID, c.ID)
			}
		}
	}
	return nil
}

// Load lee un manifiesto JSON y lo valida. Rechaza campos desconocidos y
// cualquier contenido despues del primer objeto.
func Load(r io.Reader) (Manifest, error) {
	var m Manifest
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&m); err != nil {
		return Manifest{}, fmt.Errorf("manifiesto: %w", err)
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return Manifest{}, fmt.Errorf("manifiesto: contenido extra después del primer objeto JSON")
	}
	if err := m.Validate(); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

// Dimensiones admitidas en la primera comparacion (D2). El escritorio queda fuera.
var (
	knownClients    = map[Client]bool{ClientClaude: true, ClientCodex: true}
	knownOS         = map[OS]bool{OSWindows: true, OSLinux: true}
	knownInterfaces = map[Interface]bool{InterfaceTerminal: true, InterfaceIDE: true}
	knownModalities = map[Modality]bool{ModalityAPI: true, ModalitySubscription: true}
)

// checkDimensions devuelve un motivo por dimension vacia o fuera de la matriz.
// Se resuelve como celda incompatible (no como error de Validate) porque el
// spec trata la interfaz o configuracion que no cumple como incompatible.
func checkDimensions(c Cell) []string {
	var reasons []string
	if !knownClients[c.Client] {
		reasons = append(reasons, "cliente fuera de la matriz o vacío: "+strconv.Quote(string(c.Client)))
	}
	switch {
	case c.Interface == "desktop":
		reasons = append(reasons, "interfaz de escritorio fuera de la primera comparación")
	case !knownInterfaces[c.Interface]:
		reasons = append(reasons, "interfaz fuera de la matriz o vacía: "+strconv.Quote(string(c.Interface)))
	}
	if !knownOS[c.OS] {
		reasons = append(reasons, "sistema fuera de la matriz o vacío: "+strconv.Quote(string(c.OS)))
	}
	if !knownModalities[c.Modality] {
		reasons = append(reasons, "modalidad fuera de la matriz o vacía: "+strconv.Quote(string(c.Modality)))
	}
	return reasons
}

// realInterface es la interfaz de la cohorte: una CLI que corre en la
// terminal integrada de un IDE pertenece a la cohorte terminal (spec,
// "CLI en la terminal integrada del IDE").
func realInterface(c Cell) Interface {
	if c.Observed.CLIInIDETerminal {
		return InterfaceTerminal
	}
	return c.Interface
}

// isContentOptIn reconoce la familia de opt-ins de contenido (design.md,
// "Datos sensibles"): variables OTEL_LOG_* de Claude Code y las claves de
// contenido de [otel] de Codex en forma plana o de tabla. El nombre se
// compara sin distinguir mayusculas ni espacios de los extremos.
//
// Excepcion: OTEL_LOG_LEVEL (en cualquier grafia normalizada) solo regula la
// verbosidad del propio SDK de OpenTelemetry y no registra contenido, asi que
// no cuenta (verify-2-breaker, R1). Cualquier otro OTEL_LOG_* sigue contando
// como opt-in: la familia falla cerrada.
func isContentOptIn(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "otel_log_level" {
		return false
	}
	return strings.HasPrefix(n, "otel_log_") || n == "log_user_prompt" || n == "otel.log_user_prompt"
}

// checkContentOptIns devuelve un motivo por opt-in activo, ordenado por
// nombre. Solo un valor explicitamente apagado lo desactiva; cualquier otro
// cuenta como activo.
func checkContentOptIns(settings map[string]string) []string {
	names := make([]string, 0, len(settings))
	for name := range settings {
		names = append(names, name)
	}
	sort.Strings(names)
	var reasons []string
	for _, name := range names {
		if !isContentOptIn(name) {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(settings[name])) {
		case "", "0", "false", "no", "off":
			continue
		}
		reasons = append(reasons, "opt-in de contenido activo: "+strings.TrimSpace(name))
	}
	return reasons
}

// Variantes conocidas; Validate acepta cualquier identificador no vacio.
const (
	VariantCurrent = "actual"
	VariantModular = "modular"
)
