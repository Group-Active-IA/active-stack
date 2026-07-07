using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
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

    public ComponentsPageViewModel(InstallerSessionState session, InstallSelection selection)
        : base("Choose your components", "Select the optional pieces Active Stack should include.")
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
    }

    public ObservableCollection<ComponentOption> Choices { get; }

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
            isSelected: isSelected);
        option.SelectionChanged += OnSelectionChanged;
        return option;
    }

    private void OnSelectionChanged() => SyncSelection();

    private void SyncSelection()
    {
        _selection.CustomIds = Choices.Where(static c => c.IsSelected).Select(static c => c.Id).ToList();
    }
}
