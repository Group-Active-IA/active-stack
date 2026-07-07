using System.ComponentModel;

namespace ActiveStack.Bootstrapper.Host.Pages.Uninstall;

/// <summary>A single checkbox row on the UninstallAgents page.</summary>
public sealed class AgentChoiceOption : INotifyPropertyChanged
{
    private bool _isSelected;

    public AgentChoiceOption(string id, string label, bool isSelected)
    {
        Id = id;
        Label = label;
        _isSelected = isSelected;
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    /// <summary>Raised whenever <see cref="IsSelected"/> changes, so the owning page can re-sync <c>UninstallSelection.Agents</c>.</summary>
    public event Action? SelectionChanged;

    public string Id { get; }

    public string Label { get; }

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
