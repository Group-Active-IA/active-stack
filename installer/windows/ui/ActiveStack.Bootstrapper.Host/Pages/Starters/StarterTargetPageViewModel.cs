using System.Collections.ObjectModel;
using System.Windows.Input;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages.Starters;

/// <summary>
/// Second Starter-flow page: obtains the target project directory through
/// an injectable <see cref="IFolderPicker"/> (D6, design.md) and presents
/// the agents as checkboxes defaulting to claude and opencode. Requires a
/// non-empty directory and at least one selected agent to advance. The
/// chosen path is carried as a single value (never split) so paths with
/// spaces survive.
/// </summary>
public sealed class StarterTargetPageViewModel : WizardPageViewModelBase
{
    private readonly StarterSelection _selection;
    private readonly IFolderPicker _folderPicker;
    private string _projectPath;

    public StarterTargetPageViewModel(StarterSelection selection, IFolderPicker folderPicker)
        : base("Choose a target folder", "Pick the project Active Stack should scaffold the starter into.")
    {
        _selection = selection;
        _folderPicker = folderPicker;
        _projectPath = selection.ProjectPath;

        var hasExistingSelection = selection.Agents.Count > 0;
        var preselected = hasExistingSelection ? selection.Agents : ["claude", "opencode"];

        var claude = new AgentChoiceOption("claude", "Claude", preselected.Contains("claude", StringComparer.OrdinalIgnoreCase));
        var opencode = new AgentChoiceOption("opencode", "OpenCode", preselected.Contains("opencode", StringComparer.OrdinalIgnoreCase));
        claude.SelectionChanged += OnAgentSelectionChanged;
        opencode.SelectionChanged += OnAgentSelectionChanged;

        Agents = [claude, opencode];
        SyncAgents();

        PickFolderCommand = new RelayCommand(PickFolder);
    }

    public IReadOnlyList<AgentChoiceOption> Agents { get; }

    /// <summary>Bound by the view's "Browse…" button — no code-behind click handler needed.</summary>
    public ICommand PickFolderCommand { get; }

    public string ProjectPath
    {
        get => _projectPath;
        private set
        {
            if (string.Equals(_projectPath, value, StringComparison.Ordinal))
            {
                return;
            }

            _projectPath = value;
            _selection.ProjectPath = value;
            RaisePropertyChanged(nameof(ProjectPath));
            RaiseCanAdvanceChanged();
        }
    }

    public override bool CanAdvance => !string.IsNullOrEmpty(_projectPath) && Agents.Any(static a => a.IsSelected);

    public void PickFolder()
    {
        var picked = _folderPicker.PickFolder();
        if (!string.IsNullOrEmpty(picked))
        {
            ProjectPath = picked;
        }
    }

    private void OnAgentSelectionChanged()
    {
        SyncAgents();
        RaiseCanAdvanceChanged();
    }

    private void SyncAgents()
    {
        _selection.Agents = Agents.Where(static a => a.IsSelected).Select(static a => a.Id).ToList();
    }
}
