using System.Collections.ObjectModel;

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

    public HubPageViewModel()
        : base("Active Stack", "Choose what you'd like to do.")
    {
        Entries =
        [
            new HubEntry("install", "Install", IsEnabled: true, Tooltip: null),
            new HubEntry("starters", "Starters", IsEnabled: true, Tooltip: null),
            new HubEntry("backups", "Manage Backups", IsEnabled: true, Tooltip: null),
            new HubEntry("uninstall", "Uninstall", IsEnabled: true, Tooltip: null),
            new HubEntry("update", "Update Stack — Coming soon", IsEnabled: false, Tooltip: "Coming soon")
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
