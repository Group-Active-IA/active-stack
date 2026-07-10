using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Install;

/// <summary>
/// Second Install-flow page: lite/full/custom cards with end-user labels
/// (Quick/Complete/Custom) sourced as-is from the engine's options response,
/// defaulting to the session's recommended mode ("full"). Always advanceable
/// — one mode is always selected (D4, design.md).
/// </summary>
public sealed class InstallTypePageViewModel : WizardPageViewModelBase
{
    private readonly InstallSelection _selection;
    private string _selectedId;

    public InstallTypePageViewModel(InstallerSessionState session, InstallSelection selection, string lang = "en")
        : base(UiStrings.Get(lang, "page.installtype.title"), UiStrings.Get(lang, "page.installtype.subtitle"), lang)
    {
        _selection = selection;
        Choices = new ObservableCollection<InstallTypeChoice>(session.InstallTypeChoices);

        // Honor an already-populated selection (e.g. Back navigation) instead
        // of resetting to the recommended mode every time this page is
        // (re)constructed.
        var hasExistingSelection = !string.IsNullOrEmpty(selection.Mode) &&
            Choices.Any(c => string.Equals(c.Id, selection.Mode, StringComparison.OrdinalIgnoreCase));

        _selectedId = hasExistingSelection ? selection.Mode : session.RecommendedModeId ?? "full";
        _selection.Mode = _selectedId;
    }

    public ObservableCollection<InstallTypeChoice> Choices { get; }

    public string SelectedId
    {
        get => _selectedId;
        set
        {
            if (string.Equals(_selectedId, value, StringComparison.Ordinal))
            {
                return;
            }

            _selectedId = value;
            _selection.Mode = value;
            RaisePropertyChanged(nameof(SelectedId));
            RaisePropertyChanged(nameof(DetailTitle));
            RaisePropertyChanged(nameof(DetailBody));
        }
    }

    /// <summary>The selected install type's label, for the shared detail panel.</summary>
    public string DetailTitle => SelectedChoice?.Label ?? string.Empty;

    /// <summary>The selected install type's long description, falling back to its short description (D5, design.md).</summary>
    public string DetailBody => !string.IsNullOrEmpty(SelectedChoice?.LongDescription)
        ? SelectedChoice!.LongDescription
        : SelectedChoice?.Description ?? string.Empty;

    private InstallTypeChoice? SelectedChoice => Choices.FirstOrDefault(c => string.Equals(c.Id, SelectedId, StringComparison.Ordinal));

    public override bool CanAdvance => true;
}
