using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Uninstall;

/// <summary>
/// First Uninstall-flow page: multi-select checkboxes for every agent the
/// uninstall options report as detected, pre-checked. Requires at least one
/// selection to advance — a deliberate divergence from the TUI's implicit
/// "empty means all" (D8, design.md).
/// </summary>
public sealed class UninstallAgentsPageViewModel : WizardPageViewModelBase
{
    private readonly UninstallSelection _selection;

    public UninstallAgentsPageViewModel(UninstallOptions options, UninstallSelection selection, string lang = "en")
        : base(UiStrings.Get(lang, "page.uninstallagents.title"), UiStrings.Get(lang, "page.uninstallagents.subtitle"), lang)
    {
        _selection = selection;

        var hasExistingSelection = selection.Agents.Count > 0;

        var choices = options.DetectedAgents.Select(id =>
        {
            var option = new AgentChoiceOption(id, id, hasExistingSelection ? selection.Agents.Contains(id) : true);
            option.SelectionChanged += OnSelectionChanged;
            return option;
        });

        Choices = new ObservableCollection<AgentChoiceOption>(choices);
        SyncSelection();
    }

    public ObservableCollection<AgentChoiceOption> Choices { get; }

    public override bool CanAdvance => Choices.Any(static c => c.IsSelected);

    private void OnSelectionChanged()
    {
        SyncSelection();
        RaiseCanAdvanceChanged();
    }

    private void SyncSelection()
    {
        _selection.Agents = Choices.Where(static c => c.IsSelected).Select(static c => c.Id).ToList();
    }
}
