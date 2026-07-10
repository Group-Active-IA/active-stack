using System.ComponentModel;
using System.Runtime.CompilerServices;
using ActiveStack.Bootstrapper.Core.Localization;
using ActiveStack.Bootstrapper.Host.Navigation;

namespace ActiveStack.Bootstrapper.Host.Pages;

/// <summary>
/// Shared <see cref="IWizardPage"/> plumbing: a fixed title/subtitle plus
/// property-change notification for the derived page's <c>CanAdvance</c>.
/// Keeps every concrete page view-model free of INotifyPropertyChanged
/// boilerplate. Also carries the localized <see cref="DetailHeader"/> used
/// by the shared detail-panel template (gui-detail-panel, L5, design.md D4)
/// — every page threads its <c>lang</c> constructor parameter through here
/// regardless of whether it renders a panel.
/// </summary>
public abstract class WizardPageViewModelBase : IWizardPage
{
    protected WizardPageViewModelBase(string title, string subtitle, string lang)
    {
        Title = title;
        Subtitle = subtitle;
        DetailHeader = UiStrings.Get(lang, "detail.header");
    }

    public event PropertyChangedEventHandler? PropertyChanged;

    public string Title { get; }

    public string Subtitle { get; }

    /// <summary>Localized header for the shared side detail panel (e.g. "Details"/"Detalle").</summary>
    public string DetailHeader { get; }

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
