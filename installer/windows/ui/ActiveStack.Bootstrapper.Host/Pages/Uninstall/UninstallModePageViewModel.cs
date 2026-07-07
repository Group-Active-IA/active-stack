using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
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

    public UninstallModePageViewModel(UninstallOptions options, UninstallSelection selection)
        : base("Choose uninstall mode", "Pick how much of the workspace Active Stack should remove.")
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
        }
    }

    public override bool CanAdvance => true;
}
