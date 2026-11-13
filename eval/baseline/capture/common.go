// Package capture contiene los adaptadores de captura de Claude Code y Codex
// (tareas 2.2 y 2.3 de opsx-baseline). Parsean los formatos documentados hacia
// los tipos de usage, trace y run; no ejecutan ningun cliente.
//
// Todo lo que la documentacion oficial no confirma queda "no disponible" o
// "no comprobada" en vez de inventarse. Las fuentes consultadas y sus citas
// estan en eval/baseline/README.md y en fixtures/samples/README.md.
package capture

import (
	"fmt"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
)

// Roles de la atribucion (design.md, "Consumo y reintentos").
const (
	RolePrincipal    = "principal"
	RoleWorker       = "worker"
	RoleAuxiliary    = "auxiliar"
	RoleUnattributed = "sin_atribuir"
)

// Availability dice si una fuente o atribucion esta disponible y por que no.
type Availability struct {
	Available bool
	Reason    string
}

// Status es el estado de una senal en una celda cliente/interfaz/version.
type Status string

// Estados de una senal.
const (
	StatusDocumented  Status = "documentada"
	StatusNotVerified Status = "no_comprobada"
	StatusUnavailable Status = "no_disponible"
)

// Signals es el estado de cada senal de captura para una celda.
type Signals struct {
	TokenUsage          Status
	PerAgentAttribution Status
	ToolCalls           Status
	FileReads           Status
	Permissions         Status
	Timing              Status
	Notes               []string
}

// SignalsFor devuelve el estado de las senales para un cliente, interfaz y
// version. Devuelve error si el cliente o la interfaz estan fuera de la matriz
// o si la version de Claude no es interpretable.
func SignalsFor(c manifest.Client, i manifest.Interface, version string) (Signals, error) {
	if i != manifest.InterfaceTerminal && i != manifest.InterfaceIDE {
		return Signals{}, fmt.Errorf("interfaz %q fuera de la matriz", i)
	}
	switch c {
	case manifest.ClientClaude:
		return claudeSignals(i, version)
	case manifest.ClientCodex:
		return codexSignals(i), nil
	}
	return Signals{}, fmt.Errorf("cliente %q fuera de la matriz o sin adaptador", c)
}

var claudeIDENotVerified = "la documentación de la extensión de VS Code no describe telemetría OpenTelemetry ni resultados estructurados: ninguna fuente está comprobada con la extensión real"

func claudeSignals(i manifest.Interface, version string) (Signals, error) {
	ok, err := manifest.VersionAtLeast(version, ClaudeMinAgentIDVersion)
	if err != nil {
		return Signals{}, fmt.Errorf("claude: %w", err)
	}
	if i == manifest.InterfaceIDE {
		return Signals{
			TokenUsage: StatusNotVerified, PerAgentAttribution: StatusNotVerified, ToolCalls: StatusNotVerified,
			FileReads: StatusNotVerified, Permissions: StatusNotVerified, Timing: StatusNotVerified,
			Notes: []string{claudeIDENotVerified},
		}, nil
	}
	s := Signals{
		TokenUsage: StatusDocumented, PerAgentAttribution: StatusDocumented, ToolCalls: StatusDocumented,
		FileReads: StatusUnavailable, Permissions: StatusDocumented, Timing: StatusDocumented,
		Notes: []string{
			"usage del resultado excluye a los subagentes; modelUsage y total_cost_usd los incluyen",
			"el total_tokens del cierre de un subagente cubre solo su última solicitud",
			"agent.name = custom para agentes propios: no atribuye",
			"las rutas de lectura solo con OTEL_LOG_TOOL_DETAILS=1 (opt-in prohibido en corridas medidas)",
			"la categoría subagente puede incluir hooks basados en agentes, sin evento de cierre",
		},
	}
	if !ok {
		s.PerAgentAttribution = StatusUnavailable
		s.Notes = append(s.Notes, "agent_id solo desde "+ClaudeMinAgentIDVersion)
	}
	return s, nil
}

func codexSignals(i manifest.Interface) Signals {
	if i == manifest.InterfaceIDE {
		return Signals{
			TokenUsage: StatusNotVerified, PerAgentAttribution: StatusNotVerified, ToolCalls: StatusNotVerified,
			FileReads: StatusNotVerified, Permissions: StatusNotVerified, Timing: StatusNotVerified,
			Notes: []string{"la documentación de la extensión de IDE no describe fuentes de uso ni de eventos; comparte config.toml con la CLI, pero las señales no están comprobadas con la extensión real"},
		}
	}
	return Signals{
		TokenUsage: StatusDocumented, PerAgentAttribution: StatusNotVerified, ToolCalls: StatusDocumented,
		FileReads: StatusUnavailable, Permissions: StatusNotVerified, Timing: StatusUnavailable,
		Notes: []string{
			"codex exec --json documenta el uso solo en turn.completed (entrada, entrada cacheada, salida, razonamiento); la entrada cacheada forma parte de la entrada",
			"la documentación no dice si el agregado del padre incluye a los hilos hijos ni cómo aparecen los hijos en exec --json: no comprobado, no se suma padre + hijos",
			"App Server documenta usage updates del hilo activo (thread/tokenUsage/updated) y parentThreadId en thread/list, pero no el contenido del payload de uso: no hay adaptador",
			"los eventos de turno de exec no documentan marcas de tiempo; las esperas de aprobación son de App Server",
			"las lecturas de archivos no son items propios (se hacen con command_execution)",
			"turn.failed y error se nombran pero sus campos no están documentados: se conservan crudos",
		},
	}
}
