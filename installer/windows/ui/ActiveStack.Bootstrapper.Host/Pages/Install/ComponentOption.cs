using System.ComponentModel;

namespace ActiveStack.Bootstrapper.Host.Pages.Install;

/// <summary>
/// A single checkbox row on the Components page. Forced options
/// (<see cref="IsForced"/>, i.e. the security-first "permissions" harness)
/// ignore attempts to uncheck them — mirrors the TUI's space-toggle ignore
/// for <c>install.SecurityFirstHarnessID</c> (D5, design.md).
/// </summary>
public sealed class ComponentOption : INotifyPropertyChanged
{
    private bool _isSelected;

    public ComponentOption(string id, string label, string description, bool isRecommended, bool isForced, bool isSelected)
    {
        Id = id;
        Label = label;
        Description = description;
        IsRecommended = isRecommended;
        IsForced = isForced;
        _isSelected = isSelected;
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public event Action? SelectionChanged;

    public string Id { get; }

    public string Label { get; }

    public string Description { get; }

    public bool IsRecommended { get; }

    public bool IsForced { get; }

    /// <summary>Pure negation of <see cref="IsForced"/>, exposed for XAML `IsEnabled` binding on the checkbox.</summary>
    public bool IsToggleable => !IsForced;

    public bool IsSelected
    {
        get => _isSelected;
        set
        {
            if (IsForced)
            {
                // Security-first harness: toggle is ignored, always stays checked.
                return;
            }

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
