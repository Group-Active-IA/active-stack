using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Install;

/// <summary>
/// Shown only when at least one selected agent is tier-capable (D2,
/// design.md). Single-select radio group over the session's
/// <c>PermissionTierChoices</c>, defaulting to the tier flagged
/// <c>IsDefault</c> (balanceado), surfacing that tier's warning text when
/// present (e.g. bypass).
/// </summary>
public sealed class PermissionsPageViewModel : WizardPageViewModelBase
{
    private readonly InstallSelection _selection;
    private string _selectedTierId;

    public PermissionsPageViewModel(InstallerSessionState session, InstallSelection selection, string lang = "en")
        : base(UiStrings.Get(lang, "page.permissions.title"), UiStrings.Get(lang, "page.permissions.subtitle"), lang)
    {
        _selection = selection;
        Tiers = new ObservableCollection<PermissionTierChoice>(session.PermissionTierChoices);

        // Honor an already-populated selection (e.g. Back navigation)
        // instead of resetting to the default tier every time this page is
        // (re)constructed.
        var hasExistingSelection = !string.IsNullOrEmpty(selection.Tier) &&
            Tiers.Any(t => string.Equals(t.Id, selection.Tier, StringComparison.OrdinalIgnoreCase));

        var defaultTier = Tiers.FirstOrDefault(static t => t.IsDefault) ?? Tiers.FirstOrDefault();
        _selectedTierId = hasExistingSelection ? selection.Tier! : defaultTier?.Id ?? string.Empty;
        _selection.Tier = _selectedTierId;
    }

    public ObservableCollection<PermissionTierChoice> Tiers { get; }

    public string SelectedTierId
    {
        get => _selectedTierId;
        set
        {
            if (string.Equals(_selectedTierId, value, StringComparison.Ordinal))
            {
                return;
            }

            _selectedTierId = value;
            _selection.Tier = value;
            RaisePropertyChanged(nameof(SelectedTierId));
            RaisePropertyChanged(nameof(WarningText));
            RaisePropertyChanged(nameof(DetailTitle));
            RaisePropertyChanged(nameof(DetailBody));
        }
    }

    public string? WarningText => SelectedTier?.Warning;

    /// <summary>The selected tier's label, for the shared detail panel.</summary>
    public string DetailTitle => SelectedTier?.Label ?? string.Empty;

    /// <summary>The selected tier's long description, falling back to its short description (D5, design.md).</summary>
    public string DetailBody => !string.IsNullOrEmpty(SelectedTier?.LongDescription)
        ? SelectedTier!.LongDescription
        : SelectedTier?.Description ?? string.Empty;

    private PermissionTierChoice? SelectedTier => Tiers.FirstOrDefault(t => string.Equals(t.Id, SelectedTierId, StringComparison.Ordinal));

    public override bool CanAdvance => !string.IsNullOrWhiteSpace(SelectedTierId);
}
