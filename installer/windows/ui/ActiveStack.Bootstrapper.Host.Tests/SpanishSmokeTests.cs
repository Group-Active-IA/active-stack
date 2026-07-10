using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Host.Navigation;
using ActiveStack.Bootstrapper.Host.Pages.Install;
using Xunit;

namespace ActiveStack.Bootstrapper.Host.Tests;

/// <summary>
/// Task 9.2 (gui-language-page, L4) — the Spanish smoke test, now runnable
/// since L1 (i18n-engine-locales) and L2 (catalog-localized-descriptions)
/// are merged. Verifies end-to-end CONSISTENCY at unit-test level: when the
/// shell's active language is "es", the GUI's own chrome (page
/// titles/subtitles/headings, sourced from <c>UiStrings</c>) and the
/// engine-sourced data (mode/tier/component labels, sourced from real
/// <c>--lang es</c> JSON, captured verbatim from
/// <c>go run ./cmd/active-stack windows options --lang es</c>) render in the
/// SAME language on the SAME page — no leftover English, no mixed-language
/// screens, except the one documented Go-side (L2) gap noted below.
/// </summary>
public sealed class SpanishSmokeTests
{
    private const string DetectJson = """{"detected_agents":["claude","opencode","gemini","cursor","vscode","codex","antigravity","windsurf"]}""";

    // Verbatim from `go run ./cmd/active-stack windows options --agent claude --lang es`.
    private const string OptionsJsonEs = """
    {"modes":[{"id":"lite","label":"Rápido","description":"Instalación rápida para empezar a trabajar de inmediato."},{"id":"full","label":"Completo","description":"Instalación completa recomendada con todas las herramientas clave."},{"id":"custom","label":"Personalizado","description":"Elegí exactamente qué instalar."}],"forced_components":[{"id":"permissions","label":"Permissions (security-first)","description":"Protección básica para una configuración más segura. Esto siempre se instala.","recommended":true}],"custom_components":[{"id":"openspec","label":"OpenSpec CLI","description":"CLI de Spec-Driven Development; fuente de verdad del estado.","recommended":true}],"tier_capable":true,"tier_capable_agents":["claude"],"permission_tiers":[{"id":"estricto","label":"Estricto","description":"El agente debe pedir permiso para cada operación. Máxima fricción, máxima seguridad.","default":false},{"id":"balanceado","label":"Balanceado","description":"Lista de permitidos curada para operaciones seguras y repetitivas. Punto de partida recomendado.","default":true},{"id":"bypass","label":"Bypass","description":"Autonomía total opcional. El piso de seguridad de denegación sigue aplicando.","default":false,"warning":"Bypass: modo autónomo — el piso de seguridad sigue aplicando (C-21)"}]}
    """;

    [Fact]
    public void ComponentsPage_SpanishLanguage_ChromeAndEngineComponentLabelsAreBothSpanish()
    {
        var session = InstallerSessionStateBuilder.BuildFromJson(DetectJson, OptionsJsonEs);
        var page = new ComponentsPageViewModel(session, new InstallSelection(), lang: "es");

        // Chrome: comes from UiStrings("es", ...).
        Assert.Equal("Elegí tus componentes", page.Title);
        Assert.Equal("Seleccioná las piezas opcionales que Active Stack debe incluir.", page.Subtitle);

        // Engine data: comes from the real --lang es JSON.
        var openspec = page.Choices.Single(c => c.Id == "openspec");
        Assert.Equal("OpenSpec CLI", openspec.Label);
        Assert.Contains("Spec-Driven Development", openspec.Description);

        // Documented gap (Go-side, L2, out of scope here): the forced
        // "permissions" component's LABEL is not yet localized, while its
        // description is. The C# page renders it verbatim either way — this
        // assertion pins the known gap so a future L2 fix is visible as a
        // test change here, not a silent surprise.
        var permissions = page.Choices.Single(c => c.Id == "permissions");
        Assert.Equal("Permissions (security-first)", permissions.Label);
        Assert.Contains("Protección básica", permissions.Description);
    }

    [Fact]
    public void PermissionsPage_SpanishLanguage_ChromeAndEngineTierLabelsAndWarningAreBothSpanish()
    {
        var session = InstallerSessionStateBuilder.BuildFromJson(DetectJson, OptionsJsonEs);
        var page = new PermissionsPageViewModel(session, new InstallSelection(), lang: "es");

        Assert.Equal("Elegí tu nivel de permisos", page.Title);
        Assert.Equal("Elegí con cuánta autonomía pueden actuar los agentes de Active Stack.", page.Subtitle);

        Assert.Equal("balanceado", page.SelectedTierId); // default tier, ID stable across languages

        page.SelectedTierId = "bypass";
        Assert.Contains("piso de seguridad", page.WarningText); // engine-supplied Spanish warning, shown verbatim
    }

    [Fact]
    public void ReviewPage_SpanishLanguage_HeadingsAndPrimaryLabelAreLocalized_SummariesReflectEngineData()
    {
        var session = InstallerSessionStateBuilder.BuildFromJson(DetectJson, OptionsJsonEs);
        var selection = new InstallSelection
        {
            Agents = ["claude"],
            Mode = "custom",
            CustomIds = ["permissions", "openspec"],
            Tier = "balanceado"
        };
        var page = new ReviewPageViewModel(session, selection, new NullEngineClient(), lang: "es");

        // Section headings: UiStrings("es", ...).
        Assert.Equal("Asistentes", page.AssistantsHeading);
        Assert.Equal("Tipo de instalación", page.InstallTypeHeading);
        Assert.Equal("Componentes", page.ComponentsHeading);
        Assert.Equal("Nivel de permisos", page.PermissionTierHeading);

        // Primary label: UiStrings("es", "shell.install").
        Assert.Equal("Instalar", page.PrimaryLabel);

        // Summaries: engine-sourced Spanish labels, not GUI chrome.
        Assert.Equal("Personalizado", page.ModeSummary);
        Assert.Equal("Balanceado", page.TierSummary);
        Assert.Contains("OpenSpec CLI", page.ComponentsSummary);
    }

    private sealed class NullEngineClient : IInstallerEngineClient
    {
        public string Language { get; set; } = "en";
        public Task<InstallerSessionState> LoadSessionAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunInstallAsync(IReadOnlyList<string> agents, string mode, IReadOnlyList<string> customIds, string? tier, CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<IReadOnlyList<StarterChoice>> ListStartersAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunStarterInstallAsync(string starterId, string projectPath, IReadOnlyList<string> agents, bool dryRun, CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<IReadOnlyList<BackupEntry>> ListBackupsAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<BackupActionResult> RunBackupActionAsync(string action, string id, string? description, CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public Task<UninstallOptions> LoadUninstallOptionsAsync(CancellationToken cancellationToken = default) => throw new NotSupportedException();
        public IAsyncEnumerable<InstallProgressSnapshot> RunUninstallAsync(IReadOnlyList<string> agents, string mode, string strategy, string? backupId, CancellationToken cancellationToken = default) => throw new NotSupportedException();
    }
}
