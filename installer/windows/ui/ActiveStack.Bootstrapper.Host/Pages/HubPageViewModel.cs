using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core.Localization;

namespace ActiveStack.Bootstrapper.Host.Pages;

/// <summary>
/// Entry menu of the wizard. Install, Starters, Manage Backups, and
/// Uninstall are all enabled and route to their respective flow; only
/// "Update Stack" stays disabled with a "Coming soon" tooltip, matching the
/// TUI's permanent placeholder treatment (C5 flips <c>IsEnabled</c> on the
/// three entries C4 left disabled, D8 design.md of windows-gui-hub-flows).
/// Selecting a disabled entry is a no-op: it neither changes the selection
/// nor allows advancing.
/// </summary>
public sealed class HubPageViewModel : WizardPageViewModelBase
{
    private string? _selectedEntryId;

    public HubPageViewModel(string lang = "en")
        : base(UiStrings.Get(lang, "page.hub.title"), UiStrings.Get(lang, "page.hub.subtitle"), lang)
    {
        Entries =
        [
            new HubEntry("install", UiStrings.Get(lang, "hub.entry.install"), IsEnabled: true, Tooltip: null),
            new HubEntry("starters", UiStrings.Get(lang, "hub.entry.starters"), IsEnabled: true, Tooltip: null),
            new HubEntry("backups", UiStrings.Get(lang, "hub.entry.backups"), IsEnabled: true, Tooltip: null),
            new HubEntry("uninstall", UiStrings.Get(lang, "hub.entry.uninstall"), IsEnabled: true, Tooltip: null),
            new HubEntry("update", UiStrings.Get(lang, "hub.entry.update"), IsEnabled: false, Tooltip: UiStrings.Get(lang, "hub.entry.update.tooltip"))
        ];
    }

    public ObservableCollection<HubEntry> Entries { get; }

    public string? SelectedEntryId
    {
        get => _selectedEntryId;
        set
        {
            var entry = Entries.FirstOrDefault(e => string.Equals(e.Id, value, StringComparison.Ordinal));
            if (entry is null || !entry.IsEnabled)
            {
                // Disabled (or unknown) entries cannot be selected/navigated — no-op.
                return;
            }

            if (string.Equals(_selectedEntryId, value, StringComparison.Ordinal))
            {
                return;
            }

            _selectedEntryId = value;
            RaisePropertyChanged(nameof(SelectedEntryId));
            RaiseCanAdvanceChanged();
        }
    }

    public override bool CanAdvance => _selectedEntryId is not null;
}
