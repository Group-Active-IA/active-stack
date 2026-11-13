package clientsrc

import (
	"fmt"
	"strings"
)

// Estados de una celda de la matriz (design.md D22).
const (
	StateDistributed = "distribuida"   // el cliente se instala ahí
	StateDeclared    = "declarada"     // la documentación dice que la fuente existe
	StateProbe       = "sonda"         // coincidió con la salida conocida, con generadores propios
	StateProbed      = "probada"       // la tarea 8.1 la corrió con los fixtures y dejó id de corrida
	StateTruncated   = "truncada"      // la corrida mostró recorte
	StateUnavailable = "no disponible" // la fuente no existe en esa interfaz
	StateNotProbed   = "no probada"    // con su motivo
)

var validStates = map[string]bool{
	StateDistributed: true, StateDeclared: true, StateProbe: true, StateProbed: true,
	StateTruncated: true, StateUnavailable: true, StateNotProbed: true,
}

// Entry es una fila de la matriz: una fuente en una celda cliente × interfaz × sistema.
type Entry struct {
	Client    string
	Interface string // cli | ide
	System    string // windows | linux
	Source    string
	State     string
	// RunID es el id de una corrida del grupo 8 (`8.1/...`); obligatorio para `probada`.
	RunID string
	// Backing cita lo que respalda el estado (X7, C1–C5, documentación).
	Backing string
	// CLIOnly marca una fuente que solo existe en CLI.
	CLIOnly bool
	Reason  string
}

// Validate comprueba la regla de la entrada.
func (e Entry) Validate() error {
	if !validStates[e.State] {
		return fmt.Errorf("clientsrc: estado %q desconocido en %s · %s · %s (%s)", e.State, e.Client, e.Interface, e.System, e.Source)
	}
	if e.State == StateProbed {
		if e.RunID == "" {
			return fmt.Errorf("clientsrc: una celda probada necesita el id de corrida (%s · %s · %s, %s)", e.Client, e.Interface, e.System, e.Source)
		}
		if !strings.HasPrefix(e.RunID, "8.") {
			return fmt.Errorf("clientsrc: el id de corrida %q no es del grupo 8 (%s · %s · %s)", e.RunID, e.Client, e.Interface, e.System)
		}
	}
	return nil
}

// Independent dice si la celda cuenta como captura independiente (D20): solo
// una celda probada por una corrida del grupo 8.
func (e Entry) Independent() bool { return e.State == StateProbed && e.RunID != "" }

// Matrix es la matriz de cobertura.
type Matrix struct{ Entries []Entry }

// Cell devuelve las entradas de una celda.
func (m Matrix) Cell(client, iface, system string) []Entry {
	var out []Entry
	for _, e := range m.Entries {
		if e.Client == client && e.Interface == iface && e.System == system {
			out = append(out, e)
		}
	}
	return out
}

// Validate comprueba cada entrada y que cada celda exigida (Claude y Codex, CLI
// e IDE, Windows y Linux) tenga al menos una entrada.
func (m Matrix) Validate() []string {
	var errs []string
	for _, e := range m.Entries {
		if err := e.Validate(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	for _, c := range []string{"claude", "codex"} {
		for _, i := range []string{"cli", "ide"} {
			for _, s := range []string{"windows", "linux"} {
				if len(m.Cell(c, i, s)) == 0 {
					errs = append(errs, fmt.Sprintf("celda sin entrada: %s · %s · %s", c, i, s))
				}
			}
		}
	}
	return errs
}

// DeclaredMatrix es la matriz de D22 al 2026-11-13 (entrega-03-04.md §5 y §6).
// Ninguna celda figura probada: la tarea 8.1 todavía no corrió.
func DeclaredMatrix() Matrix {
	var es []Entry
	add := func(e Entry) { es = append(es, e) }

	const noLinux = "sin cliente nativo en WSL (entrega-03-04.md §6)"
	for _, c := range []string{"claude", "codex"} {
		for _, i := range []string{"cli", "ide"} {
			add(Entry{Client: c, Interface: i, System: "linux", Source: "todas las fuentes", State: StateNotProbed, Reason: noLinux})
		}
	}

	// Claude Code · Windows.
	add(Entry{Client: "claude", Interface: "cli", System: "windows", Source: "tool_result de un comando fallido (stream-json)", State: StateTruncated,
		Backing: "C1–C5: ~10.040 caracteres de cabeza y cola; la cola no es el final; sin ruta"})
	add(Entry{Client: "claude", Interface: "cli", System: "windows", Source: "archivo del cliente en un comando exitoso (tool-results)", State: StateProbe,
		Backing: "C6: solo exit 0, copia sin redactar"})
	add(Entry{Client: "claude", Interface: "cli", System: "windows", Source: "hooks PostToolUse / PostToolUseFailure", State: StateDeclared,
		Backing: "https://code.claude.com/docs/en/hooks; no se instalaron hooks"})
	add(Entry{Client: "claude", Interface: "ide", System: "windows", Source: "hooks PostToolUse / PostToolUseFailure", State: StateDeclared,
		Backing: "https://code.claude.com/docs/en/hooks (\"fires the same hook events wherever it runs\"); sin IDE operable"})

	// Codex · Windows.
	add(Entry{Client: "codex", Interface: "cli", System: "windows", Source: "codex exec --json: aggregated_output", State: StateProbe,
		Backing: "X7", CLIOnly: true})
	add(Entry{Client: "codex", Interface: "ide", System: "windows", Source: "codex exec --json: aggregated_output", State: StateUnavailable,
		Backing: "solo existe en CLI", CLIOnly: true, Reason: "el arnés no conduce `codex exec` dentro de la extensión"})
	add(Entry{Client: "codex", Interface: "cli", System: "windows", Source: "lo que recibe el modelo; historial de la sesión", State: StateTruncated,
		Backing: "X1–X6, X8: cabeza y cola sin aviso; el historial guarda la recortada"})
	add(Entry{Client: "codex", Interface: "cli", System: "windows", Source: "App Server: outputDelta y aggregatedOutput", State: StateDeclared,
		Backing: "https://learn.chatgpt.com/docs/app-server"})
	add(Entry{Client: "codex", Interface: "ide", System: "windows", Source: "App Server interno de la extensión", State: StateUnavailable,
		Backing: "documentado", Reason: "el arnés no puede suscribirse a la instancia de la extensión"})
	return Matrix{Entries: es}
}
