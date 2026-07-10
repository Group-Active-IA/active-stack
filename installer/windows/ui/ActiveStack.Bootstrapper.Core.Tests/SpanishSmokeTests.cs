using Xunit;

namespace ActiveStack.Bootstrapper.Core.Tests;

/// <summary>
/// Task 9.2 (gui-language-page, L4): now that L1 (i18n-engine-locales) and
/// L2 (catalog-localized-descriptions) are merged, the Go engine's
/// <c>--lang es</c> output is real. These fixtures are VERBATIM captures
/// from <c>go run ./cmd/active-stack windows options|starters list|uninstall-options --lang es</c>
/// (and <c>windows detect</c>, which never carries <c>--lang</c> per the
/// frozen contract) — not hand-written guesses — so parsing them proves the
/// Core parsers correctly pass real engine-localized text through to the
/// GUI's data model without any English-text coupling.
/// </summary>
public sealed class SpanishSmokeTests
{
    private const string DetectJson = """{"detected_agents":["claude","opencode","gemini","cursor","vscode","codex","antigravity","windsurf"]}""";

    private const string OptionsJsonEs = """
    {"modes":[{"id":"lite","label":"Rápido","description":"Instalación rápida para empezar a trabajar de inmediato.","long_description":"El modo Rápido instala lo esencial para que puedas empezar a trabajar de inmediato. Omite los extras opcionales y las pantallas de configuración avanzada. Siempre podés agregar más después con el modo Personalizado."},{"id":"full","label":"Completo","description":"Instalación completa recomendada con todas las herramientas clave.","long_description":"El modo Completo instala todos los harnesses y herramientas recomendadas para una configuración totalmente equipada. Tarda un poco más que el modo Rápido, pero no deja afuera nada importante. Ideal cuando querés todo el kit desde el primer día."},{"id":"custom","label":"Personalizado","description":"Elegí exactamente qué instalar.","long_description":"El modo Personalizado te permite elegir exactamente qué componentes instalar. Usalo cuando ya conocés tu stack y querés control total. No se instala nada fuera de tu selección (además de la base de seguridad obligatoria)."}],"forced_components":[{"id":"permissions","label":"Permissions (security-first)","description":"Protección básica para una configuración más segura. Esto siempre se instala.","recommended":true,"long_description":"Preconfigura las reglas de permisos del agente para bloquear archivos sensibles y pedir confirmación antes de operaciones git destructivas."}],"custom_components":[{"id":"openspec","label":"OpenSpec CLI","description":"CLI de Spec-Driven Development; fuente de verdad del estado.","recommended":true,"long_description":"Gestiona el ciclo explorar → proponer → aplicar → archivar de los changes del proyecto, manteniendo specs y tareas sincronizadas."}],"tier_capable":true,"tier_capable_agents":["claude"],"permission_tiers":[{"id":"estricto","label":"Estricto","description":"El agente debe pedir permiso para cada operación. Máxima fricción, máxima seguridad.","default":false,"long_description":"Estricto exige que el agente pida confirmación antes de cada operación."},{"id":"balanceado","label":"Balanceado","description":"Lista de permitidos curada para operaciones seguras y repetitivas. Punto de partida recomendado.","default":true,"long_description":"Balanceado usa una lista de permitidos curada."},{"id":"bypass","label":"Bypass","description":"Autonomía total opcional. El piso de seguridad de denegación sigue aplicando.","default":false,"warning":"Bypass: modo autónomo — el piso de seguridad sigue aplicando (C-21)","long_description":"Bypass le da al agente autonomía total para actuar sin preguntar."}]}
    """;

    private const string StartersJsonEs = """
    {"starters":[{"id":"base","name":"Base","description":"Transversales metodológicos compartidos por todos los starters (TDD, debugging, code review, producto ágil).","includes":[],"harnesses":["test-driven-development","systematic-debugging"],"mcp_count":0,"long_description":"Bundle base de 6 harnesses transversales."},{"id":"backend","name":"Backend","description":"Stack de backend Python/FastAPI con arquitectura limpia, BD, seguridad y contenedores.","includes":["base"],"harnesses":["clean-architecture","fastapi-domain-service"],"mcp_count":0,"long_description":"Compone base + 12 harnesses de capa backend."}]}
    """;

    private const string UninstallOptionsJsonEs = """
    {"detected_agents":["claude","opencode","gemini","cursor","vscode","codex","antigravity","windsurf"],"modes":[{"id":"lite","label":"Rápido","description":"Instalación rápida para empezar a trabajar de inmediato."},{"id":"full","label":"Completo","description":"Instalación completa recomendada con todas las herramientas clave."},{"id":"custom","label":"Personalizado","description":"Elegí exactamente qué instalar."}],"strategies":[{"id":"targeted","label":"Dirigida","description":"Revierte cada harness instalado de forma individual.","default":true,"requires_manifest":false},{"id":"restore","label":"Restaurar desde backup","description":"Restaura el estado previo completo desde un manifiesto de backup.","default":false,"requires_manifest":true}]}
    """;

    [Fact]
    public void InstallerSessionStateBuilder_ParsesRealSpanishOptionsResponse_LabelsAreSpanishIdsAreContractStable()
    {
        var state = InstallerSessionStateBuilder.BuildFromJson(DetectJson, OptionsJsonEs);

        // Contract IDs (frozen, D-language) must NOT change under --lang es.
        Assert.Contains(state.InstallTypeChoices, m => m.Id == "lite");
        Assert.Contains(state.InstallTypeChoices, m => m.Id == "full");
        Assert.Contains(state.InstallTypeChoices, m => m.Id == "custom");
        Assert.Contains(state.PermissionTierChoices, t => t.Id == "balanceado");
        Assert.Contains(state.PermissionTierChoices, t => t.Id == "bypass");

        // Labels/descriptions must be the real Spanish text from the engine.
        Assert.Equal("Rápido", state.InstallTypeChoices.Single(m => m.Id == "lite").Label);
        Assert.Equal("Completo", state.InstallTypeChoices.Single(m => m.Id == "full").Label);
        Assert.Equal("Personalizado", state.InstallTypeChoices.Single(m => m.Id == "custom").Label);
        Assert.Equal("Balanceado", state.PermissionTierChoices.Single(t => t.Id == "balanceado").Label);
        Assert.Contains("piso de seguridad", state.PermissionTierChoices.Single(t => t.Id == "bypass").Warning);
        Assert.Equal("OpenSpec CLI", state.CustomComponents.Single(c => c.Id == "openspec").Label);
        Assert.Contains("Spec-Driven Development", state.CustomComponents.Single(c => c.Id == "openspec").Description);

        // recommendedModeId selection is ID-based, unaffected by language.
        Assert.Equal("full", state.RecommendedModeId);
    }

    [Fact]
    public void StarterCatalogParser_ParsesRealSpanishStartersResponse()
    {
        var starters = StarterCatalogParser.BuildFromJson(StartersJsonEs);

        Assert.Equal(2, starters.Count);
        Assert.Equal("Base", starters[0].Name);
        Assert.Contains("Transversales metodológicos", starters[0].Description);
        Assert.Equal("Backend", starters[1].Name);
        Assert.Contains("arquitectura limpia", starters[1].Description);
    }

    [Fact]
    public void UninstallOptionsParser_ParsesRealSpanishUninstallOptionsResponse()
    {
        var options = UninstallOptionsParser.BuildFromJson(UninstallOptionsJsonEs);

        Assert.Contains(options.Modes, m => m.Id == "lite" && m.Label == "Rápido");
        Assert.Contains(options.Strategies, s => s.Id == "targeted" && s.Label == "Dirigida");
        Assert.Contains(options.Strategies, s => s.Id == "restore" && s.Label == "Restaurar desde backup");

        // Contract IDs stable under --lang es.
        Assert.True(options.Strategies.Single(s => s.Id == "targeted").IsDefault);
        Assert.True(options.Strategies.Single(s => s.Id == "restore").RequiresManifest);
    }

    [Fact]
    public void ForcedComponentLabel_IsAKnownGoSideL2Gap_NotAParsingBug()
    {
        // Documents a real finding from the L1/L2 smoke: the engine's
        // forced "permissions" component label is not yet localized
        // ("Permissions (security-first)" ships in es responses too).
        // This is a Go-side catalog-data gap (out of scope for this C#
        // change) — the C# parser correctly passes it through verbatim
        // rather than mangling or silently dropping it.
        var state = InstallerSessionStateBuilder.BuildFromJson(DetectJson, OptionsJsonEs);

        var permissions = state.ForcedComponents.Single(c => c.Id == "permissions");
        Assert.Equal("Permissions (security-first)", permissions.Label);
        // Its description IS localized, confirming the gap is label-only.
        Assert.Contains("Protección básica", permissions.Description);
    }
}
