using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Install;

/// <summary>
/// Custom-mode-only page: lists the session's forced + custom components as
/// checkboxes. The security-first "permissions" harness (forced component)
/// is rendered checked and disabled with the engine-provided copy and is
/// always present in the resulting custom ids (D5, design.md).
/// </summary>
public sealed class ComponentsPageViewModel : WizardPageViewModelBase
{
    private readonly InstallSelection _selection;
    private string? _selectedComponentId;

    public ComponentsPageViewModel(InstallerSessionState session, InstallSelection selection, string lang = "en")
        : base(UiStrings.Get(lang, "page.components.title"), UiStrings.Get(lang, "page.components.subtitle"), lang)
    {
        _selection = selection;

        // Honor an already-populated selection (e.g. Back navigation)
        // instead of resetting every checkbox to "recommended" every time
        // this page is (re)constructed.
        var hasExistingSelection = selection.CustomIds.Count > 0;

        var options = new List<ComponentOption>();

        foreach (var forced in session.ForcedComponents)
        {
            options.Add(CreateOption(forced, isForced: true, hasExistingSelection));
        }

        foreach (var custom in session.CustomComponents)
        {
            options.Add(CreateOption(custom, isForced: false, hasExistingSelection));
        }

        Choices = new ObservableCollection<ComponentOption>(options);
        SyncSelection();

        // Row selection (which item is focused for the detail panel) is
        // independent of each row's checkbox — defaults to the first
        // rendered option, never feeds the engine payload (D3, design.md).
        _selectedComponentId = Choices.FirstOrDefault()?.Id;
    }

    public ObservableCollection<ComponentOption> Choices { get; }

    /// <summary>
    /// The row currently focused/highlighted (feeds the detail panel).
    /// Entirely independent of each row's <c>IsSelected</c> checkbox —
    /// never derives or is derived from <see cref="InstallSelection.CustomIds"/>
    /// (D3, design.md).
    /// </summary>
    public string? SelectedComponentId
    {
        get => _selectedComponentId;
        set
        {
            if (string.Equals(_selectedComponentId, value, StringComparison.Ordinal))
            {
                return;
            }

            _selectedComponentId = value;
            RaisePropertyChanged(nameof(SelectedComponentId));
            RaisePropertyChanged(nameof(DetailTitle));
            RaisePropertyChanged(nameof(DetailBody));
        }
    }

    /// <summary>The focused row's label, for the shared detail panel.</summary>
    public string DetailTitle => SelectedOption?.Label ?? string.Empty;

    /// <summary>The focused row's long description, falling back to its short description (D5, design.md).</summary>
    public string DetailBody => !string.IsNullOrEmpty(SelectedOption?.LongDescription)
        ? SelectedOption!.LongDescription
        : SelectedOption?.Description ?? string.Empty;

    private ComponentOption? SelectedOption => Choices.FirstOrDefault(c => string.Equals(c.Id, SelectedComponentId, StringComparison.Ordinal))
        ?? Choices.FirstOrDefault();

    public override bool CanAdvance => true;

    private ComponentOption CreateOption(ComponentChoice choice, bool isForced, bool hasExistingSelection)
    {
        var isSelected = isForced || (hasExistingSelection ? _selection.CustomIds.Contains(choice.Id) : choice.Recommended);
        var option = new ComponentOption(
            choice.Id,
            choice.Label,
            choice.Description,
            isRecommended: choice.Recommended,
            isForced: isForced,
            isSelected: isSelected,
            longDescription: choice.LongDescription);
        option.SelectionChanged += OnSelectionChanged;
        return option;
    }

    private void OnSelectionChanged() => SyncSelection();

    private void SyncSelection()
    {
        _selection.CustomIds = Choices.Where(static c => c.IsSelected).Select(static c => c.Id).ToList();
    }
}
