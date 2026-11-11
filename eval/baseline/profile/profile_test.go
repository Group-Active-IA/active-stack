package profile

import (
	"strings"
	"testing"

	"github.com/Group-Active-IA/active-stack/eval/baseline/manifest"
)

func TestProfiles_CoverEveryClientInterfaceKind(t *testing.T) {
	ps := Profiles()
	if len(ps) != 8 {
		t.Fatalf("len(Profiles()) = %d, want 8 (2 clientes x 2 interfaces x 2 tipos)", len(ps))
	}
	seenID := map[string]bool{}
	for _, c := range []manifest.Client{manifest.ClientClaude, manifest.ClientCodex} {
		for _, i := range []manifest.Interface{manifest.InterfaceTerminal, manifest.InterfaceIDE} {
			for _, k := range []Kind{KindEquivalent, KindStrict} {
				p, ok := Lookup(c, i, k)
				if !ok {
					t.Fatalf("Lookup(%s,%s,%s): no hay perfil", c, i, k)
				}
				if p.Client != c || p.Interface != i || p.Kind != k {
					t.Errorf("Lookup(%s,%s,%s) devolvio %s/%s/%s", c, i, k, p.Client, p.Interface, p.Kind)
				}
				if seenID[p.ID] || p.ID == "" {
					t.Errorf("id vacio o repetido: %q", p.ID)
				}
				seenID[p.ID] = true
			}
		}
	}
	if _, ok := Lookup("gemini", manifest.InterfaceTerminal, KindStrict); ok {
		t.Error("Lookup de un cliente fuera de la matriz debe fallar")
	}
}

func TestProfiles_EveryProfileIsValidAndDocumented(t *testing.T) {
	for _, p := range Profiles() {
		if err := p.Validate(); err != nil {
			t.Errorf("%s: Validate() = %v", p.ID, err)
		}
		if len(p.Sources) == 0 {
			t.Errorf("%s: sin fuentes de documentacion", p.ID)
		}
		for _, s := range p.Sources {
			if !strings.HasPrefix(s.URL, "https://") || s.Consulted == "" || strings.TrimSpace(s.Quote) == "" {
				t.Errorf("%s: fuente incompleta %+v", p.ID, s)
			}
		}
	}
}

func TestEditInsideCheckoutNeedsApproval(t *testing.T) {
	for _, p := range Profiles() {
		got, err := p.EditInsideCheckoutNeedsApproval()
		if err != nil {
			t.Errorf("%s: error %v", p.ID, err)
			continue
		}
		want := p.Kind == KindStrict
		if got != want {
			t.Errorf("%s: EditInsideCheckoutNeedsApproval = %v, want %v", p.ID, got, want)
		}
	}
}

func TestProfiles_ConcreteModes(t *testing.T) {
	cases := []struct {
		c    manifest.Client
		i    manifest.Interface
		k    Kind
		mode string
	}{
		{manifest.ClientClaude, manifest.InterfaceTerminal, KindEquivalent, "acceptEdits"},
		{manifest.ClientClaude, manifest.InterfaceIDE, KindEquivalent, "acceptEdits"},
		{manifest.ClientClaude, manifest.InterfaceTerminal, KindStrict, "default"},
		{manifest.ClientClaude, manifest.InterfaceIDE, KindStrict, "default"},
		{manifest.ClientCodex, manifest.InterfaceTerminal, KindEquivalent, "workspace-write/on-request"},
		{manifest.ClientCodex, manifest.InterfaceIDE, KindEquivalent, "workspace-write/on-request"},
		{manifest.ClientCodex, manifest.InterfaceTerminal, KindStrict, "read-only/on-request"},
		{manifest.ClientCodex, manifest.InterfaceIDE, KindStrict, "read-only/on-request"},
	}
	for _, tc := range cases {
		p, ok := Lookup(tc.c, tc.i, tc.k)
		if !ok {
			t.Fatalf("falta perfil %s/%s/%s", tc.c, tc.i, tc.k)
		}
		if p.Permissions.Mode != tc.mode {
			t.Errorf("%s: modo %q, want %q", p.ID, p.Permissions.Mode, tc.mode)
		}
		if len(p.Permissions.Flags) == 0 {
			t.Errorf("%s: sin flags efectivos documentados", p.ID)
		}
	}
}

func claudeEquivalent(t *testing.T) Profile {
	t.Helper()
	p, ok := Lookup(manifest.ClientClaude, manifest.InterfaceTerminal, KindEquivalent)
	if !ok {
		t.Fatal("falta perfil claude/terminal/equivalente")
	}
	return p
}

func withFlags(p Profile, mode string, flags ...string) Profile {
	p.Permissions = manifest.Permissions{Mode: mode, Flags: flags}
	return p
}

func TestValidate_RejectsBroaderPermissions(t *testing.T) {
	base := claudeEquivalent(t)
	codex, _ := Lookup(manifest.ClientCodex, manifest.InterfaceTerminal, KindEquivalent)
	cases := []struct {
		name string
		p    Profile
	}{
		{"claude bypassPermissions", withFlags(base, "bypassPermissions", "--permission-mode bypassPermissions")},
		{"claude modo auto", withFlags(base, "auto", "--permission-mode auto")},
		{"claude modo desconocido", withFlags(base, "yolo", "--permission-mode yolo")},
		{"claude flag de modo distinto del modo", withFlags(base, "acceptEdits", "--permission-mode default")},
		{"claude sin flag de modo", withFlags(base, "acceptEdits", "--allowedTools Edit(./**)")},
		{"claude Edit sin alcance", withFlags(base, "acceptEdits", "--permission-mode acceptEdits", "--allowedTools Edit")},
		{"claude Bash sin alcance", withFlags(base, "acceptEdits", "--permission-mode acceptEdits", "--allowedTools Bash")},
		{"claude Edit fuera del checkout", withFlags(base, "acceptEdits", "--permission-mode acceptEdits", "--allowedTools Edit(//etc/**)")},
		{"claude Edit con home", withFlags(base, "acceptEdits", "--permission-mode acceptEdits", "--allowedTools Edit(~/**)")},
		{"claude directorio adicional", withFlags(base, "acceptEdits", "--permission-mode acceptEdits", "--add-dir ../otro")},
		{"claude saltear permisos", withFlags(base, "acceptEdits", "--permission-mode acceptEdits", "--dangerously-skip-permissions")},
		{"claude Bash con comodin", withFlags(base, "acceptEdits", "--permission-mode acceptEdits", "--allowedTools Bash(*)")},
		{"claude Bash en blanco", withFlags(base, "acceptEdits", "--permission-mode acceptEdits", "--allowedTools Bash( )")},
		{"claude modo declarado distinto del flag (Estricto)", Profile{ID: "x", Client: manifest.ClientClaude, Interface: manifest.InterfaceTerminal, Kind: KindStrict,
			Permissions: manifest.Permissions{Mode: "acceptEdits", Flags: []string{"--permission-mode default"}}}},
		{"claude equivalente que pide aprobacion al editar", Profile{ID: "x", Client: manifest.ClientClaude, Interface: manifest.InterfaceTerminal, Kind: KindEquivalent,
			Permissions: manifest.Permissions{Mode: "default", Flags: []string{"--permission-mode default"}}}},
		{"claude Estricto que no pide aprobacion al editar", Profile{ID: "x", Client: manifest.ClientClaude, Interface: manifest.InterfaceTerminal, Kind: KindStrict,
			Permissions: manifest.Permissions{Mode: "acceptEdits", Flags: []string{"--permission-mode acceptEdits"}}}},
		{"claude Estricto con una regla allow de shell", Profile{ID: "x", Client: manifest.ClientClaude, Interface: manifest.InterfaceTerminal, Kind: KindStrict,
			Permissions: manifest.Permissions{Mode: "default", Flags: []string{"--permission-mode default", "--allowedTools Bash(go test *)"}}}},
		{"codex Estricto con workspace-write", Profile{ID: "x", Client: manifest.ClientCodex, Interface: manifest.InterfaceTerminal, Kind: KindStrict,
			Permissions: manifest.Permissions{Mode: "workspace-write/on-request", Flags: []string{"--sandbox workspace-write", "--ask-for-approval on-request"}}}},
		{"codex acceso total", withFlags(codex, "danger-full-access/on-request", "--sandbox danger-full-access", "--ask-for-approval on-request")},
		{"codex modo sin politica", withFlags(codex, "workspace-write", "--sandbox workspace-write")},
		{"codex aprobacion retirada", withFlags(codex, "workspace-write/untrusted", "--sandbox workspace-write", "--ask-for-approval untrusted")},
		{"codex bypass", withFlags(codex, "workspace-write/on-request", "--sandbox workspace-write", "--ask-for-approval on-request", "--dangerously-bypass-approvals-and-sandbox")},
		{"codex flag de sandbox distinto del modo", withFlags(codex, "workspace-write/on-request", "--sandbox read-only", "--ask-for-approval on-request")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.p.Validate(); err == nil {
				t.Errorf("Validate() = nil, want error para %+v", tc.p.Permissions)
			}
		})
	}
	// Lo benigno sigue validando: Edit con alcance relativo al checkout.
	ok := withFlags(base, "acceptEdits", "--permission-mode acceptEdits", "--allowedTools Edit(./**)", "--allowedTools Bash(go test *)")
	if err := ok.Validate(); err != nil {
		t.Errorf("Validate() de un allow con alcance = %v, want nil", err)
	}
}

func TestEditInsideCheckoutNeedsApproval_UnknownModeIsAnError(t *testing.T) {
	p := withFlags(claudeEquivalent(t), "plan", "--permission-mode plan")
	if _, err := p.EditInsideCheckoutNeedsApproval(); err == nil {
		t.Error("un modo fuera de los documentados para medir debe ser error, no un si/no")
	}
	c, _ := Lookup(manifest.ClientCodex, manifest.InterfaceTerminal, KindEquivalent)
	c = withFlags(c, "workspace-write/never", "--sandbox workspace-write", "--ask-for-approval never")
	got, err := c.EditInsideCheckoutNeedsApproval()
	if err != nil || got {
		t.Errorf("workspace-write/never: (%v, %v), want (false, nil): no pide aprobacion dentro del workspace", got, err)
	}
	ro, _ := Lookup(manifest.ClientCodex, manifest.InterfaceTerminal, KindStrict)
	ro = withFlags(ro, "read-only/never", "--sandbox read-only", "--ask-for-approval never")
	got, err = ro.EditInsideCheckoutNeedsApproval()
	if err != nil || !got {
		t.Errorf("read-only/never: (%v, %v), want (true, nil): la edicion queda bloqueada, no aprobada", got, err)
	}
}

func TestEditInsideCheckoutNeedsApproval_ManualModeWithScopedEditAllow(t *testing.T) {
	p := Profile{ID: "x", Client: manifest.ClientClaude, Interface: manifest.InterfaceTerminal, Kind: KindEquivalent,
		Permissions: manifest.Permissions{Mode: "default", Flags: []string{"--permission-mode default", "--allowedTools Edit(./**)"}}}
	got, err := p.EditInsideCheckoutNeedsApproval()
	if err != nil || got {
		t.Errorf("Manual con Edit(./**): (%v, %v), want (false, nil): la regla allow evita la aprobacion", got, err)
	}
	if err := p.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
	// Un Read con alcance no evita la aprobacion de editar.
	p.Permissions.Flags = []string{"--permission-mode default", "--allowedTools Read(./**)"}
	if got, _ := p.EditInsideCheckoutNeedsApproval(); !got {
		t.Error("Manual con solo Read(./**): editar debe pedir aprobacion")
	}
}

func TestCompare_ObservedAgainstProfile(t *testing.T) {
	p := claudeEquivalent(t)
	same := manifest.Observed{Permissions: manifest.Permissions{Mode: p.Permissions.Mode, Flags: reversed(p.Permissions.Flags)}}
	if got := Compare(p, same); got.Status != manifest.CellExecutable {
		t.Errorf("observado igual (otro orden): %+v, want ejecutable", got)
	}
	cases := []struct {
		name string
		obs  manifest.Observed
		want string
	}{
		{"modo distinto", manifest.Observed{Permissions: manifest.Permissions{Mode: "default", Flags: p.Permissions.Flags}}, "modo default en lugar de acceptEdits"},
		{"flag extra", manifest.Observed{Permissions: manifest.Permissions{Mode: p.Permissions.Mode, Flags: append(append([]string{}, p.Permissions.Flags...), "--add-dir ..")}}, "flags observados"},
		{"flag faltante", manifest.Observed{Permissions: manifest.Permissions{Mode: p.Permissions.Mode, Flags: p.Permissions.Flags[:len(p.Permissions.Flags)-1]}}, "flags observados"},
		{"opt-in de contenido de Claude", manifest.Observed{Permissions: p.Permissions, Settings: map[string]string{"OTEL_LOG_USER_PROMPTS": "1"}}, "OTEL_LOG_USER_PROMPTS"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Compare(p, tc.obs)
			if got.Status != manifest.CellIncompatible {
				t.Fatalf("Compare = %+v, want incompatible", got)
			}
			if !strings.Contains(strings.Join(got.Reasons, "|"), tc.want) {
				t.Errorf("motivos = %v, want que incluya %q", got.Reasons, tc.want)
			}
		})
	}
	// Una serie Estricto observada con los permisos del perfil equivalente no pasa.
	strict, _ := Lookup(manifest.ClientCodex, manifest.InterfaceTerminal, KindStrict)
	eq, _ := Lookup(manifest.ClientCodex, manifest.InterfaceTerminal, KindEquivalent)
	if got := Compare(strict, manifest.Observed{Permissions: eq.Permissions}); got.Status != manifest.CellIncompatible {
		t.Errorf("Estricto con permisos de equivalente: %+v, want incompatible", got)
	}
	if got := Compare(strict, manifest.Observed{Permissions: strict.Permissions}); got.Status != manifest.CellExecutable {
		t.Errorf("Estricto con sus permisos: %+v, want ejecutable", got)
	}
	effective := strict.Permissions
	effective.Mode = strings.Replace(effective.Mode, "on-request", "never", 1)
	if effective.Mode == strict.Permissions.Mode {
		t.Fatal("strict profile has no on-request policy")
	}
	if got := Compare(strict, manifest.Observed{Permissions: effective}); got.Status != manifest.CellIncompatible {
		t.Errorf("Codex exec sin escalación: %+v, want incompatible", got)
	}
	// Opt-in de Codex.
	if got := Compare(eq, manifest.Observed{Permissions: eq.Permissions, Settings: map[string]string{"otel.log_user_prompt": "true"}}); got.Status != manifest.CellIncompatible {
		t.Errorf("log_user_prompt activo: %+v, want incompatible", got)
	}
}

func reversed(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[len(in)-1-i] = s
	}
	return out
}

func TestCellMatches(t *testing.T) {
	p := claudeEquivalent(t)
	cell := manifest.Cell{Client: p.Client, Interface: p.Interface, Permissions: p.Permissions}
	if r := CellMatches(p, cell); len(r) != 0 {
		t.Errorf("celda igual al perfil: motivos %v, want ninguno", r)
	}
	cases := []struct {
		name string
		mut  func(c *manifest.Cell)
		want string
	}{
		{"otro cliente", func(c *manifest.Cell) { c.Client = manifest.ClientCodex }, "cliente"},
		{"otra interfaz", func(c *manifest.Cell) { c.Interface = manifest.InterfaceIDE }, "interfaz"},
		{"otro modo", func(c *manifest.Cell) { c.Permissions.Mode = "default" }, "modo"},
		{"otros flags", func(c *manifest.Cell) { c.Permissions.Flags = []string{"--permission-mode acceptEdits"} }, "flags"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := cell
			c.Permissions = manifest.Permissions{Mode: p.Permissions.Mode, Flags: append([]string{}, p.Permissions.Flags...)}
			tc.mut(&c)
			r := CellMatches(p, c)
			if len(r) == 0 || !strings.Contains(strings.Join(r, "|"), tc.want) {
				t.Errorf("motivos = %v, want que incluya %q", r, tc.want)
			}
		})
	}
}

func TestPending_RealVerificationOnWindowsAndLinuxForEveryProfile(t *testing.T) {
	items := Pending()
	got := map[string]map[manifest.OS]bool{}
	for _, it := range items {
		if got[it.ProfileID] == nil {
			got[it.ProfileID] = map[manifest.OS]bool{}
		}
		got[it.ProfileID][it.OS] = true
		if strings.TrimSpace(it.Check) == "" {
			t.Errorf("%s/%s: pendiente sin descripcion de la comprobacion", it.ProfileID, it.OS)
		}
	}
	for _, p := range Profiles() {
		for _, os := range []manifest.OS{manifest.OSWindows, manifest.OSLinux} {
			if !got[p.ID][os] {
				t.Errorf("%s en %s: falta la verificacion real pendiente", p.ID, os)
			}
		}
	}
}
