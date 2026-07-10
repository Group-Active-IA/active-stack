using System.Collections.ObjectModel;
using ActiveStack.Bootstrapper.Core;
using ActiveStack.Bootstrapper.Core.Localization;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Install;

/// <summary>
/// First Install-flow page: multi-select checkboxes for every agent the
/// session reports, pre-checked (agents surfaced by "windows detect" are, by
/// construction, the detected ones) with the session's default agent marked
/// recommended. Requires at least one selection to advance (D4, design.md).
/// </summary>
public sealed class AssistantsPageViewModel : WizardPageViewModelBase
{
    private readonly InstallSelection _selection;

    public AssistantsPageViewModel(InstallerSessionState session, InstallSelection selection, string lang = "en")
        : base(UiStrings.Get(lang, "page.assistants.title"), UiStrings.Get(lang, "page.assistants.subtitle"), lang)
    {
        _selection = selection;
        RecommendedLabel = UiStrings.Get(lang, "template.recommended");

        // Honor an already-populated selection (e.g. the user navigated Back
        // to this page) instead of resetting to "every agent checked" —
        // only pre-check everything on the page's first-ever visit, when
        // the shared selection has not been written to yet.
        var hasExistingSelection = selection.Agents.Count > 0;

        var options = session.AssistantChoices.Select(assistant =>
        {
            var option = new AssistantChoiceOption(
                assistant.Id,
                assistant.Label,
                isRecommended: string.Equals(assistant.Id, session.DefaultAssistantId, StringComparison.OrdinalIgnoreCase),
                isSelected: hasExistingSelection ? selection.Agents.Contains(assistant.Id) : true);
            option.SelectionChanged += OnSelectionChanged;
            return option;
        });

        Choices = new ObservableCollection<AssistantChoiceOption>(options);
        SyncSelection();
    }

    public ObservableCollection<AssistantChoiceOption> Choices { get; }

    /// <summary>Localized "Recommended" tag rendered next to the default assistant (bound from the item template).</summary>
    public string RecommendedLabel { get; }

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
