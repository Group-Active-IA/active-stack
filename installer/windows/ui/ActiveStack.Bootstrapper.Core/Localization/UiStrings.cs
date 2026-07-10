namespace ActiveStack.Bootstrapper.Core.Localization;

/// <summary>
/// Pure, static UI string table for the Windows GUI Host. <see cref="Get"/>
/// has NO mutable global state — the caller (the shell) owns the current
/// language and passes it in on every call. Covers the ~85 Host literals:
/// shell footer actions, every page's title/subtitle, hub entries, review
/// headings, PageTemplates headers, backups/uninstall warnings, the
/// complete-state labels, and the progress format strings. Falls back
/// en -&gt; key when a translation is missing (gui-language-page, L4).
/// NOT implemented with .resx/satellite assemblies: those break
/// <c>PublishSingleFile</c> in build.ps1.
/// </summary>
public static class UiStrings
{
    public static string Get(string lang, string key)
    {
        var table = TableFor(lang);
        if (table.TryGetValue(key, out var value) && !string.IsNullOrEmpty(value))
        {
            return value;
        }

        if (En.TryGetValue(key, out var fallback) && !string.IsNullOrEmpty(fallback))
        {
            return fallback;
        }

        return key;
    }

    /// <summary>Exposed for the en/es parity test (Core.Tests) — not for production use.</summary>
    public static IReadOnlyCollection<string> EnglishKeys => En.Keys;

    /// <summary>Exposed for the en/es parity test (Core.Tests) — not for production use.</summary>
    public static IReadOnlyCollection<string> SpanishKeys => Es.Keys;

    private static Dictionary<string, string> TableFor(string lang) =>
        string.Equals(lang, "es", StringComparison.OrdinalIgnoreCase) ? Es : En;

    private static readonly Dictionary<string, string> En = new()
    {
        // Shell footer + stream-trigger primary labels.
        ["shell.next"] = "Next",
        ["shell.finish"] = "Finish",
        ["shell.back"] = "Back",
        ["shell.cancel"] = "Cancel",
        ["shell.install"] = "Install",
        ["shell.uninstall"] = "Uninstall",

        // Language page (fixed bilingual copy — identical in both tables).
        ["page.language.title"] = "Select your language / Elegí tu idioma",
        ["page.language.subtitle"] = "Choose the language Active Stack should use. / Elegí el idioma que va a usar Active Stack.",
        ["page.language.choice.english"] = "English",
        ["page.language.choice.spanish"] = "Español",

        // Page titles/subtitles.
        ["page.hub.title"] = "Active Stack",
        ["page.hub.subtitle"] = "Choose what you'd like to do.",
        ["page.assistants.title"] = "Choose your assistants",
        ["page.assistants.subtitle"] = "Select every coding assistant Active Stack should set up.",
        ["page.installtype.title"] = "Choose your installation type",
        ["page.installtype.subtitle"] = "Pick how much of the workspace Active Stack should set up.",
        ["page.components.title"] = "Choose your components",
        ["page.components.subtitle"] = "Select the optional pieces Active Stack should include.",
        ["page.permissions.title"] = "Choose your permission tier",
        ["page.permissions.subtitle"] = "Pick how autonomously Active Stack's agents can act.",
        ["page.review.title"] = "Review your setup",
        ["page.review.subtitle"] = "Confirm the selection before Active Stack installs it.",
        ["page.installing.title"] = "Installing Active Stack",
        ["page.installing.subtitle"] = "Sit tight while Active Stack sets up your workspace.",
        ["page.complete.subtitle"] = "Here's how your Active Stack setup went.",
        ["page.uninstallagents.title"] = "Choose agents to uninstall",
        ["page.uninstallagents.subtitle"] = "Select every coding assistant Active Stack should remove.",
        ["page.uninstallmode.title"] = "Choose uninstall mode",
        ["page.uninstallmode.subtitle"] = "Pick how much of the workspace Active Stack should remove.",
        ["page.uninstallstrategy.title"] = "Choose uninstall strategy",
        ["page.uninstallstrategy.subtitle"] = "Pick whether to remove everything or restore a previous backup.",
        ["page.uninstallconfirm.title"] = "Review the uninstall",
        ["page.uninstallconfirm.subtitle"] = "Confirm the selection before Active Stack removes it.",
        ["page.startercatalog.title"] = "Choose a starter",
        ["page.startercatalog.subtitle"] = "Pick a starter to scaffold into your project.",
        ["page.startertarget.title"] = "Choose a target folder",
        ["page.startertarget.subtitle"] = "Pick the project Active Stack should scaffold the starter into.",
        ["page.starterreview.title"] = "Review your starter",
        ["page.starterreview.subtitle"] = "Confirm the selection before Active Stack scaffolds it.",
        ["page.backups.title"] = "Manage backups",
        ["page.backups.subtitle"] = "Restore, rename, or delete a previous Active Stack backup.",

        // Hub entries + tooltips.
        ["hub.entry.install"] = "Install",
        ["hub.entry.starters"] = "Starters",
        ["hub.entry.backups"] = "Manage Backups",
        ["hub.entry.uninstall"] = "Uninstall",
        ["hub.entry.update"] = "Update Stack — Coming soon",
        ["hub.entry.update.tooltip"] = "Coming soon",

        // Install Review headings.
        ["review.heading.assistants"] = "Assistants",
        ["review.heading.installtype"] = "Install type",
        ["review.heading.components"] = "Components",
        ["review.heading.permissiontier"] = "Permission tier",

        // Uninstall Confirm headings.
        ["uninstallreview.heading.agents"] = "Agents",
        ["uninstallreview.heading.mode"] = "Mode",
        ["uninstallreview.heading.strategy"] = "Strategy",
        ["uninstallreview.heading.restoringbackup"] = "Restoring backup",

        // Starter Review headings.
        ["starterreview.heading.starter"] = "Starter",
        ["starterreview.heading.targetfolder"] = "Target folder",
        ["starterreview.heading.agents"] = "Agents",

        // Starter Target page headings.
        ["startertarget.heading.target"] = "Target project",
        ["startertarget.heading.agents"] = "Agents",

        // PageTemplates.xaml shared chrome.
        ["template.recommended"] = "Recommended",
        ["template.choosebackuptorestore"] = "Choose a backup to restore",
        ["template.nobackupsyet"] = "No backups yet.",
        ["template.restore"] = "Restore",
        ["template.rename"] = "Rename",
        ["template.delete"] = "Delete",
        ["template.confirm"] = "Confirm",
        ["template.cancel"] = "Cancel",
        ["template.browse"] = "Browse…",
        ["template.harnesses_fmt"] = "{0} harnesses",

        // Backups warnings.
        ["backups.warning.overwrite"] = "This will OVERWRITE your current configuration.",
        ["backups.warning.delete"] = "This will PERMANENTLY DELETE this backup.",

        // Uninstall modification warning.
        ["uninstall.warning"] = "This will modify your agent configuration.",

        // Complete page: state titles + short state labels.
        ["complete.state.success.title"] = "Installation complete",
        ["complete.state.degraded.title"] = "Installation complete (with some best-effort steps skipped)",
        ["complete.state.rolledback.title"] = "Installation rolled back",
        ["complete.state.error.title"] = "Installation failed",
        ["complete.state.success.label"] = "Success",
        ["complete.state.degraded.label"] = "Degraded",
        ["complete.state.rolledback.label"] = "Rolled back",
        ["complete.state.error.label"] = "Error",

        // Progress step messages / format strings.
        ["progress.installing_fmt"] = "Installing {0}.",
        ["progress.installed_fmt"] = "Installed {0}.",
        ["progress.failed_fmt"] = "Failed to install {0}.",
        ["progress.downloading_fmt"] = "Downloading {0}.",
        ["progress.downloaded_fmt"] = "Downloaded {0}.",
        ["progress.running_default"] = "Running installation.",
        ["progress.finished_success_default"] = "Installation finished successfully.",
        ["progress.finished_failed_default"] = "Installation failed.",

        // Detail panel (gui-detail-panel, L5): shared header for the 6 pages
        // that render a selected item's long description in a side panel.
        ["detail.header"] = "Details",
    };

    private static readonly Dictionary<string, string> Es = new()
    {
        ["shell.next"] = "Siguiente",
        ["shell.finish"] = "Finalizar",
        ["shell.back"] = "Atrás",
        ["shell.cancel"] = "Cancelar",
        ["shell.install"] = "Instalar",
        ["shell.uninstall"] = "Desinstalar",

        ["page.language.title"] = "Select your language / Elegí tu idioma",
        ["page.language.subtitle"] = "Choose the language Active Stack should use. / Elegí el idioma que va a usar Active Stack.",
        ["page.language.choice.english"] = "English",
        ["page.language.choice.spanish"] = "Español",

        ["page.hub.title"] = "Active Stack",
        ["page.hub.subtitle"] = "Elegí qué te gustaría hacer.",
        ["page.assistants.title"] = "Elegí tus asistentes",
        ["page.assistants.subtitle"] = "Seleccioná cada asistente de código que Active Stack debe configurar.",
        ["page.installtype.title"] = "Elegí tu tipo de instalación",
        ["page.installtype.subtitle"] = "Elegí cuánto del espacio de trabajo debe configurar Active Stack.",
        ["page.components.title"] = "Elegí tus componentes",
        ["page.components.subtitle"] = "Seleccioná las piezas opcionales que Active Stack debe incluir.",
        ["page.permissions.title"] = "Elegí tu nivel de permisos",
        ["page.permissions.subtitle"] = "Elegí con cuánta autonomía pueden actuar los agentes de Active Stack.",
        ["page.review.title"] = "Revisá tu configuración",
        ["page.review.subtitle"] = "Confirmá la selección antes de que Active Stack la instale.",
        ["page.installing.title"] = "Instalando Active Stack",
        ["page.installing.subtitle"] = "Esperá mientras Active Stack configura tu espacio de trabajo.",
        ["page.complete.subtitle"] = "Así resultó la configuración de tu Active Stack.",
        ["page.uninstallagents.title"] = "Elegí los agentes a desinstalar",
        ["page.uninstallagents.subtitle"] = "Seleccioná cada asistente de código que Active Stack debe eliminar.",
        ["page.uninstallmode.title"] = "Elegí el modo de desinstalación",
        ["page.uninstallmode.subtitle"] = "Elegí cuánto del espacio de trabajo debe eliminar Active Stack.",
        ["page.uninstallstrategy.title"] = "Elegí la estrategia de desinstalación",
        ["page.uninstallstrategy.subtitle"] = "Elegí si eliminar todo o restaurar un backup anterior.",
        ["page.uninstallconfirm.title"] = "Revisá la desinstalación",
        ["page.uninstallconfirm.subtitle"] = "Confirmá la selección antes de que Active Stack la elimine.",
        ["page.startercatalog.title"] = "Elegí un starter",
        ["page.startercatalog.subtitle"] = "Elegí un starter para incorporar a tu proyecto.",
        ["page.startertarget.title"] = "Elegí una carpeta de destino",
        ["page.startertarget.subtitle"] = "Elegí el proyecto donde Active Stack debe incorporar el starter.",
        ["page.starterreview.title"] = "Revisá tu starter",
        ["page.starterreview.subtitle"] = "Confirmá la selección antes de que Active Stack lo incorpore.",
        ["page.backups.title"] = "Gestioná backups",
        ["page.backups.subtitle"] = "Restaurá, renombrá o eliminá un backup anterior de Active Stack.",

        ["hub.entry.install"] = "Instalar",
        ["hub.entry.starters"] = "Starters",
        ["hub.entry.backups"] = "Gestionar backups",
        ["hub.entry.uninstall"] = "Desinstalar",
        ["hub.entry.update"] = "Actualizar Stack — Próximamente",
        ["hub.entry.update.tooltip"] = "Próximamente",

        ["review.heading.assistants"] = "Asistentes",
        ["review.heading.installtype"] = "Tipo de instalación",
        ["review.heading.components"] = "Componentes",
        ["review.heading.permissiontier"] = "Nivel de permisos",

        ["uninstallreview.heading.agents"] = "Agentes",
        ["uninstallreview.heading.mode"] = "Modo",
        ["uninstallreview.heading.strategy"] = "Estrategia",
        ["uninstallreview.heading.restoringbackup"] = "Restaurando backup",

        ["starterreview.heading.starter"] = "Starter",
        ["starterreview.heading.targetfolder"] = "Carpeta de destino",
        ["starterreview.heading.agents"] = "Agentes",

        ["startertarget.heading.target"] = "Proyecto de destino",
        ["startertarget.heading.agents"] = "Agentes",

        ["template.recommended"] = "Recomendado",
        ["template.choosebackuptorestore"] = "Elegí un backup para restaurar",
        ["template.nobackupsyet"] = "Todavía no hay backups.",
        ["template.restore"] = "Restaurar",
        ["template.rename"] = "Renombrar",
        ["template.delete"] = "Eliminar",
        ["template.confirm"] = "Confirmar",
        ["template.cancel"] = "Cancelar",
        ["template.browse"] = "Examinar…",
        ["template.harnesses_fmt"] = "{0} harnesses",

        ["backups.warning.overwrite"] = "Esto SOBRESCRIBIRÁ tu configuración actual.",
        ["backups.warning.delete"] = "Esto ELIMINARÁ PERMANENTEMENTE este backup.",

        ["uninstall.warning"] = "Esto modificará la configuración de tus agentes.",

        ["complete.state.success.title"] = "Instalación completa",
        ["complete.state.degraded.title"] = "Instalación completa (se omitieron algunos pasos de mejor esfuerzo)",
        ["complete.state.rolledback.title"] = "Instalación revertida",
        ["complete.state.error.title"] = "La instalación falló",
        ["complete.state.success.label"] = "Éxito",
        ["complete.state.degraded.label"] = "Con advertencias",
        ["complete.state.rolledback.label"] = "Revertida",
        ["complete.state.error.label"] = "Error",

        ["progress.installing_fmt"] = "Instalando {0}.",
        ["progress.installed_fmt"] = "Instalado {0}.",
        ["progress.failed_fmt"] = "Error al instalar {0}.",
        ["progress.downloading_fmt"] = "Descargando {0}.",
        ["progress.downloaded_fmt"] = "Descargado {0}.",
        ["progress.running_default"] = "Ejecutando instalación.",
        ["progress.finished_success_default"] = "Instalación finalizada con éxito.",
        ["progress.finished_failed_default"] = "La instalación falló.",

        ["detail.header"] = "Detalle",
    };
}
