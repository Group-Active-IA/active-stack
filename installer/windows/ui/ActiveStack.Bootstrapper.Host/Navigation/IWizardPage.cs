using System.ComponentModel;

namespace ActiveStack.Bootstrapper.Host.Navigation;

/// <summary>
/// Contract every wizard page view-model implements so the shell
/// (<see cref="ShellViewModel"/>) can render it generically: a header
/// (<see cref="Title"/>/<see cref="Subtitle"/>) and whether the shell's
/// primary footer action should be enabled (<see cref="CanAdvance"/>).
/// Pages that need to react to selection changes (e.g. the Assistants
/// min-1 gate) raise <see cref="INotifyPropertyChanged.PropertyChanged"/>
/// for <see cref="CanAdvance"/> so the shell stays in sync.
/// </summary>
public interface IWizardPage : INotifyPropertyChanged
{
    string Title { get; }

    string Subtitle { get; }

    bool CanAdvance { get; }
}
