using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Starters;

/// <summary>
/// First Starter-flow page: a single-select list of the starters returned by
/// <see cref="IInstallerEngineClient.ListStartersAsync"/>, showing each
/// starter's name, description, and harness count. Requires a selection to
/// advance.
/// </summary>
public sealed class StarterCatalogPageViewModel : WizardPageViewModelBase
{
    private readonly StarterSelection _selection;
    private string? _selectedStarterId;

    public StarterCatalogPageViewModel(IReadOnlyList<StarterChoice> starters, StarterSelection selection, string lang = "en")
        : base(UiStrings.Get(lang, "page.startercatalog.title"), UiStrings.Get(lang, "page.startercatalog.subtitle"), lang)
    {
        _selection = selection;
        Choices = new ObservableCollection<StarterChoice>(starters);

        var hasExistingSelection = !string.IsNullOrEmpty(selection.StarterId) &&
            Choices.Any(c => string.Equals(c.Id, selection.StarterId, StringComparison.OrdinalIgnoreCase));

        _selectedStarterId = hasExistingSelection ? selection.StarterId : null;
        _selection.StarterId = _selectedStarterId ?? string.Empty;
    }

    public ObservableCollection<StarterChoice> Choices { get; }

    public string? SelectedStarterId
    {
        get => _selectedStarterId;
        set
        {
            if (string.Equals(_selectedStarterId, value, StringComparison.Ordinal))
            {
                return;
            }

            _selectedStarterId = value;
            _selection.StarterId = value ?? string.Empty;
            RaisePropertyChanged(nameof(SelectedStarterId));
            RaisePropertyChanged(nameof(DetailTitle));
            RaisePropertyChanged(nameof(DetailBody));
            RaiseCanAdvanceChanged();
        }
    }

    /// <summary>The selected starter's name, for the shared detail panel.</summary>
    public string DetailTitle => SelectedStarter?.Name ?? string.Empty;

    /// <summary>The selected starter's long description, falling back to its short description (D5, design.md).</summary>
    public string DetailBody => !string.IsNullOrEmpty(SelectedStarter?.LongDescription)
        ? SelectedStarter!.LongDescription
        : SelectedStarter?.Description ?? string.Empty;

    private StarterChoice? SelectedStarter => Choices.FirstOrDefault(c => string.Equals(c.Id, _selectedStarterId, StringComparison.Ordinal));

    public override bool CanAdvance => !string.IsNullOrEmpty(_selectedStarterId);
}
