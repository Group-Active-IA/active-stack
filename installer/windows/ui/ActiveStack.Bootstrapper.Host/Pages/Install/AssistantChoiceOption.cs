using System.ComponentModel;

namespace ActiveStack.Bootstrapper.Host.Pages.Install;

/// <summary>
/// A single checkbox row on the Assistants page.
/// </summary>
public sealed class AssistantChoiceOption : INotifyPropertyChanged
{
    private bool _isSelected;

    public AssistantChoiceOption(string id, string label, bool isRecommended, bool isSelected)
    {
        Id = id;
        Label = label;
        IsRecommended = isRecommended;
        _isSelected = isSelected;
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    /// <summary>Raised whenever <see cref="IsSelected"/> changes, so the owning page can re-sync <c>InstallSelection.Agents</c>.</summary>
    public event Action? SelectionChanged;

    public string Id { get; }

    public string Label { get; }

    public bool IsRecommended { get; }

    public bool IsSelected
    {
        get => _isSelected;
        set
        {
            if (_isSelected == value)
            {
                return;
            }

            _isSelected = value;
            PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(nameof(IsSelected)));
            SelectionChanged?.Invoke();
        }
    }
}
