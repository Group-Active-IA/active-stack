using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;

namespace ActiveStack.Bootstrapper.Host.Pages.Install;

/// <summary>
/// Shown between Hub and Assistants in the Install flow (windows-gui-wizard,
/// MODIFIED "Install flow ordering mirrors the TUI"). Read-only report of the
/// engine's dependency detection — mirrors the TUI's ScreenDetection. Never
/// blocks: CanAdvance is always true regardless of missing required
/// dependencies (gui-installer-dependency-detection spec, "Dependencies page
/// never blocks advancing" — any real failure surfaces later through the
/// install engine's own pipeline, unchanged by this page).
/// </summary>
public sealed class DependenciesPageViewModel : WizardPageViewModelBase
{
    public DependenciesPageViewModel(InstallerSessionState session, string lang = "en")
        : base(UiStrings.Get(lang, "page.dependencies.title"), UiStrings.Get(lang, "page.dependencies.subtitle"), lang)
    {
        var installedLabel = UiStrings.Get(lang, "dependencies.status.installed");
        var missingLabel = UiStrings.Get(lang, "dependencies.status.missing");
        var requiredLabel = UiStrings.Get(lang, "dependencies.label.required");
        var optionalLabel = UiStrings.Get(lang, "dependencies.label.optional");
        RunLabel = UiStrings.Get(lang, "dependencies.label.run");

        Rows = new ObservableCollection<DependencyRow>(session.Dependencies.Select(dep => new DependencyRow(
            dep.Name,
            dep.Required,
            dep.Installed,
            dep.Version,
            StatusLabel: dep.Installed ? installedLabel : missingLabel,
            RequiredLabel: dep.Required ? requiredLabel : optionalLabel,
            InstallHint: dep.Installed ? string.Empty : dep.InstallHint,
            InstallCommand: dep.Installed ? string.Empty : dep.InstallCommand)));
    }

    public ObservableCollection<DependencyRow> Rows { get; }

    /// <summary>Localized "Run:" label prefixed to a row's copyable install command.</summary>
    public string RunLabel { get; }

    public override bool CanAdvance => true;
}
