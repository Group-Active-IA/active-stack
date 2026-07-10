using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Uninstall;

/// <summary>
/// Second Uninstall-flow page: lite/full/custom as a single-select radio
/// group sourced from the uninstall options' modes, defaulting to "full"
/// when present. Always advanceable — one mode is always selected.
/// </summary>
public sealed class UninstallModePageViewModel : WizardPageViewModelBase
{
    private readonly UninstallSelection _selection;
    private string _selectedId;

    public UninstallModePageViewModel(UninstallOptions options, UninstallSelection selection, string lang = "en")
        : base(UiStrings.Get(lang, "page.uninstallmode.title"), UiStrings.Get(lang, "page.uninstallmode.subtitle"), lang)
    {
        _selection = selection;
        Choices = new ObservableCollection<InstallTypeChoice>(options.Modes);

        var hasExistingSelection = !string.IsNullOrEmpty(selection.Mode) &&
            Choices.Any(c => string.Equals(c.Id, selection.Mode, StringComparison.OrdinalIgnoreCase));

        var defaultMode = Choices.FirstOrDefault(static c => string.Equals(c.Id, "full", StringComparison.OrdinalIgnoreCase))
            ?? Choices.FirstOrDefault();

        _selectedId = hasExistingSelection ? selection.Mode : defaultMode?.Id ?? string.Empty;
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

    /// <summary>The selected mode's label, for the shared detail panel.</summary>
    public string DetailTitle => SelectedChoice?.Label ?? string.Empty;

    /// <summary>The selected mode's long description, falling back to its short description (D5, design.md).</summary>
    public string DetailBody => !string.IsNullOrEmpty(SelectedChoice?.LongDescription)
        ? SelectedChoice!.LongDescription
        : SelectedChoice?.Description ?? string.Empty;

    private InstallTypeChoice? SelectedChoice => Choices.FirstOrDefault(c => string.Equals(c.Id, SelectedId, StringComparison.Ordinal));

    public override bool CanAdvance => true;
}
