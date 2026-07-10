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

    public ComponentOption(string id, string label, string description, bool isRecommended, bool isForced, bool isSelected, string longDescription = "")
    {
        Id = id;
        Label = label;
        Description = description;
        IsRecommended = isRecommended;
        IsForced = isForced;
        _isSelected = isSelected;
        LongDescription = longDescription;
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public event Action? SelectionChanged;

    public string Id { get; }

    public string Label { get; }

    public string Description { get; }

    /// <summary>Long-form copy for the shared detail panel (gui-detail-panel, L5, design.md D1). Empty when the engine did not supply one.</summary>
    public string LongDescription { get; }

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
