using System.ComponentModel;
using System.Runtime.CompilerServices;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages;

/// <summary>
/// Shared <see cref="IWizardPage"/> plumbing: a fixed title/subtitle plus
/// property-change notification for the derived page's <c>CanAdvance</c>.
/// Keeps every concrete page view-model free of INotifyPropertyChanged
/// boilerplate.
/// </summary>
public abstract class WizardPageViewModelBase : IWizardPage
{
    protected WizardPageViewModelBase(string title, string subtitle)
    {
        Title = title;
        Subtitle = subtitle;
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public string Title { get; }

    public string Subtitle { get; }

    public abstract bool CanAdvance { get; }

    protected void RaisePropertyChanged(string propertyName) =>
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(propertyName));

    protected void RaiseCanAdvanceChanged() => RaisePropertyChanged(nameof(CanAdvance));

    protected void SetField<T>(ref T field, T value, [CallerMemberName] string? propertyName = null)
    {
        if (EqualityComparer<T>.Default.Equals(field, value))
        {
            return;
        }

        field = value;
        RaisePropertyChanged(propertyName!);
    }
}
